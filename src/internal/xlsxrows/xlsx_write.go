package xlsxrows

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// WriteOptions controls the small, dependency-free XLSX writer used by SGL.
// Rows are written as inline strings so the same file can be read back by Read.
type WriteOptions struct {
	SheetName    string
	InputColumns map[int]bool // zero-based columns highlighted as user input
	Widths       []float64
	FreezeHeader bool
	AutoFilter   bool
}

// Write creates a standards-compliant XLSX workbook using only the Go standard
// library. It intentionally supports the subset SGL needs for export/import.
func Write(w io.Writer, rows [][]string, opts WriteOptions) error {
	if w == nil {
		return fmt.Errorf("destino Excel no disponible")
	}
	name := strings.TrimSpace(opts.SheetName)
	if name == "" {
		name = "ASPIRANTES"
	}
	name = sanitizeSheetName(name)
	zw := zip.NewWriter(w)
	files := map[string]string{
		"[Content_Types].xml":        contentTypesXML,
		"_rels/.rels":                rootRelsXML,
		"xl/workbook.xml":            workbookXML(name),
		"xl/_rels/workbook.xml.rels": workbookRelsXML,
		"xl/styles.xml":              stylesXML,
		"xl/worksheets/sheet1.xml":   worksheetXML(rows, opts),
	}
	order := []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml", "xl/worksheets/sheet1.xml"}
	for _, path := range order {
		entry, err := zw.Create(path)
		if err != nil {
			_ = zw.Close()
			return err
		}
		if _, err := io.WriteString(entry, files[path]); err != nil {
			_ = zw.Close()
			return err
		}
	}
	return zw.Close()
}

func sanitizeSheetName(name string) string {
	replacer := strings.NewReplacer("[", " ", "]", " ", ":", " ", "*", " ", "?", " ", "/", " ", "\\", " ")
	name = strings.TrimSpace(replacer.Replace(name))
	if name == "" {
		name = "ASPIRANTES"
	}
	runes := []rune(name)
	if len(runes) > 31 {
		name = string(runes[:31])
	}
	return name
}

func xmlText(value string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}

func columnName(index int) string {
	if index < 0 {
		return "A"
	}
	index++
	out := ""
	for index > 0 {
		index--
		out = string(rune('A'+index%26)) + out
		index /= 26
	}
	return out
}

func maxColumns(rows [][]string) int {
	max := 0
	for _, row := range rows {
		if len(row) > max {
			max = len(row)
		}
	}
	return max
}

func worksheetXML(rows [][]string, opts WriteOptions) string {
	cols := maxColumns(rows)
	if cols == 0 {
		cols = 1
	}
	lastCol := columnName(cols - 1)
	lastRow := len(rows)
	if lastRow < 1 {
		lastRow = 1
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	b.WriteString(`<dimension ref="A1:` + lastCol + strconv.Itoa(lastRow) + `"/>`)
	if opts.FreezeHeader {
		b.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/><selection pane="bottomLeft" activeCell="A2" sqref="A2"/></sheetView></sheetViews>`)
	} else {
		b.WriteString(`<sheetViews><sheetView workbookViewId="0"/></sheetViews>`)
	}
	b.WriteString(`<sheetFormatPr defaultRowHeight="18"/>`)
	if cols > 0 {
		b.WriteString(`<cols>`)
		for i := 0; i < cols; i++ {
			width := 18.0
			if i < len(opts.Widths) && opts.Widths[i] > 0 {
				width = opts.Widths[i]
			}
			b.WriteString(`<col min="` + strconv.Itoa(i+1) + `" max="` + strconv.Itoa(i+1) + `" width="` + strconv.FormatFloat(width, 'f', 1, 64) + `" customWidth="1"/>`)
		}
		b.WriteString(`</cols>`)
	}
	b.WriteString(`<sheetData>`)
	for r, row := range rows {
		rowNum := r + 1
		height := "20"
		if r == 0 {
			height = "24"
		}
		b.WriteString(`<row r="` + strconv.Itoa(rowNum) + `" ht="` + height + `" customHeight="1">`)
		for c, value := range row {
			style := 2 // linked/imported data: green text
			if r == 0 {
				style = 1 // header
			} else if opts.InputColumns != nil && opts.InputColumns[c] {
				style = 3 // user input / caution
			}
			ref := columnName(c) + strconv.Itoa(rowNum)
			b.WriteString(`<c r="` + ref + `" s="` + strconv.Itoa(style) + `" t="inlineStr"><is><t xml:space="preserve">` + xmlText(value) + `</t></is></c>`)
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData>`)
	if opts.AutoFilter && len(rows) > 0 && cols > 0 {
		b.WriteString(`<autoFilter ref="A1:` + lastCol + strconv.Itoa(lastRow) + `"/>`)
	}
	b.WriteString(`<pageMargins left="0.3" right="0.3" top="0.5" bottom="0.5" header="0.2" footer="0.2"/>`)
	b.WriteString(`</worksheet>`)
	return b.String()
}

func workbookXML(sheetName string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<bookViews><workbookView xWindow="0" yWindow="0" windowWidth="24000" windowHeight="14000"/></bookViews>` +
		`<sheets><sheet name="` + xmlText(sheetName) + `" sheetId="1" r:id="rId1"/></sheets></workbook>`
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const workbookRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

// Style 1: header. Style 2: imported data (green text). Style 3: cells the
// administrator must complete/review (blue text on light orange fill).
const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<fonts count="4">
  <font><sz val="11"/><name val="Calibri"/><family val="2"/></font>
  <font><b/><color rgb="FFFFFFFF"/><sz val="11"/><name val="Calibri"/><family val="2"/></font>
  <font><color rgb="FF008000"/><sz val="11"/><name val="Calibri"/><family val="2"/></font>
  <font><color rgb="FF0000FF"/><sz val="11"/><name val="Calibri"/><family val="2"/></font>
</fonts>
<fills count="4">
  <fill><patternFill patternType="none"/></fill>
  <fill><patternFill patternType="gray125"/></fill>
  <fill><patternFill patternType="solid"><fgColor rgb="FF1F4E78"/><bgColor indexed="64"/></patternFill></fill>
  <fill><patternFill patternType="solid"><fgColor rgb="FFFCE4D6"/><bgColor indexed="64"/></patternFill></fill>
</fills>
<borders count="2">
  <border><left/><right/><top/><bottom/><diagonal/></border>
  <border><left/><right/><top/><bottom style="thin"><color rgb="FFD9E2F3"/></bottom><diagonal/></border>
</borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="4">
  <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyAlignment="1"><alignment vertical="center"/></xf>
  <xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
  <xf numFmtId="0" fontId="2" fillId="0" borderId="1" xfId="0" applyFont="1" applyBorder="1" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf>
  <xf numFmtId="0" fontId="3" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment vertical="center" wrapText="1"/></xf>
</cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
