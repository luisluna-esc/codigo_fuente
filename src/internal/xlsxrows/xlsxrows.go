package xlsxrows

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Read extracts the first worksheet of a simple XLSX file as rows of strings.
// It intentionally uses only the Go standard library so the installed app does
// not need external runtimes or Excel installed on the machine.
func Read(r io.Reader, maxBytes int64) ([][]string, error) {
	if maxBytes <= 0 {
		maxBytes = 12 << 20
	}
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("archivo Excel demasiado grande")
	}
	return fromBytes(data)
}

func fromBytes(data []byte) ([][]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("no es un XLSX valido")
	}
	shared, _ := sharedStrings(zr)
	sheetPath := firstSheetPath(zr)
	if sheetPath == "" {
		sheetPath = "xl/worksheets/sheet1.xml"
	}
	sheet, err := readZip(zr, sheetPath)
	if err != nil {
		return nil, fmt.Errorf("no se encontro la primera hoja de Excel")
	}
	return parseSheet(sheet, shared)
}

func readZip(zr *zip.Reader, name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "/")
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("no encontrado: %s", name)
}

func firstSheetPath(zr *zip.Reader) string {
	book, err := readZip(zr, "xl/workbook.xml")
	if err != nil {
		return ""
	}
	rels, _ := workbookRelationships(zr)
	dec := xml.NewDecoder(bytes.NewReader(book))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "sheet" {
			continue
		}
		id := ""
		for _, attr := range start.Attr {
			if attr.Name.Local == "id" {
				id = attr.Value
				break
			}
		}
		if target := rels[id]; target != "" {
			if strings.HasPrefix(target, "/") {
				return strings.TrimPrefix(target, "/")
			}
			if strings.HasPrefix(target, "xl/") {
				return target
			}
			return "xl/" + strings.TrimPrefix(target, "/")
		}
		return "xl/worksheets/sheet1.xml"
	}
	return ""
}

func workbookRelationships(zr *zip.Reader) (map[string]string, error) {
	data, err := readZip(zr, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}
	type rel struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	}
	type rels struct {
		Items []rel `xml:"Relationship"`
	}
	var parsed rels
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, item := range parsed.Items {
		out[item.ID] = item.Target
	}
	return out, nil
}

func sharedStrings(zr *zip.Reader) ([]string, error) {
	data, err := readZip(zr, "xl/sharedStrings.xml")
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	items := []string{}
	var builder strings.Builder
	inSI := false
	inT := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "si" {
				inSI = true
				builder.Reset()
			} else if inSI && t.Name.Local == "t" {
				inT = true
			}
		case xml.CharData:
			if inSI && inT {
				builder.Write([]byte(t))
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inT = false
			} else if t.Name.Local == "si" {
				items = append(items, builder.String())
				inSI = false
			}
		}
	}
	return items, nil
}

type cellState struct {
	ref        string
	cellType   string
	valueText  strings.Builder
	inlineText strings.Builder
}

func parseSheet(data []byte, shared []string) ([][]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	rows := [][]string{}
	rowCells := map[int]string{}
	maxCol := -1
	inRow := false
	var cell *cellState
	readMode := ""
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				inRow = true
				rowCells = map[int]string{}
				maxCol = -1
			case "c":
				cell = &cellState{}
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "r":
						cell.ref = attr.Value
					case "t":
						cell.cellType = attr.Value
					}
				}
			case "v":
				if cell != nil {
					readMode = "v"
				}
			case "t":
				if cell != nil {
					readMode = "t"
				}
			}
		case xml.CharData:
			if cell != nil {
				if readMode == "v" {
					cell.valueText.Write([]byte(t))
				} else if readMode == "t" {
					cell.inlineText.Write([]byte(t))
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v", "t":
				readMode = ""
			case "c":
				if inRow && cell != nil {
					col := columnIndex(cell.ref)
					if col < 0 {
						col = maxCol + 1
					}
					value := strings.TrimSpace(cell.valueText.String())
					if cell.cellType == "s" {
						idx, _ := strconv.Atoi(value)
						if idx >= 0 && idx < len(shared) {
							value = shared[idx]
						}
					} else if cell.cellType == "inlineStr" || cell.inlineText.Len() > 0 {
						value = cell.inlineText.String()
					}
					rowCells[col] = value
					if col > maxCol {
						maxCol = col
					}
				}
				cell = nil
			case "row":
				if maxCol >= 0 {
					row := make([]string, maxCol+1)
					nonEmpty := false
					for i := 0; i <= maxCol; i++ {
						row[i] = strings.TrimSpace(rowCells[i])
						if row[i] != "" {
							nonEmpty = true
						}
					}
					if nonEmpty {
						rows = append(rows, row)
					}
				}
				inRow = false
			}
		}
	}
	return rows, nil
}

func columnIndex(ref string) int {
	if ref == "" {
		return -1
	}
	idx := 0
	used := false
	for _, r := range ref {
		if r >= 'A' && r <= 'Z' {
			idx = idx*26 + int(r-'A'+1)
			used = true
		} else if r >= 'a' && r <= 'z' {
			idx = idx*26 + int(r-'a'+1)
			used = true
		} else {
			break
		}
	}
	if !used {
		return -1
	}
	return idx - 1
}
