package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"sgl.local/sistema-gestion-liderazgo/internal/xlsxrows"
)

type xlsxExportRequest struct {
	Sheet        string     `json:"sheet"`
	Rows         [][]string `json:"rows"`
	InputColumns []int      `json:"inputColumns"`
	Widths       []float64  `json:"widths"`
	Filename     string     `json:"filename"`
}

func (s *Server) exportXlsx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, http.StatusForbidden, "Esta ruta requiere Administrador.", nil)
		return
	}
	defer r.Body.Close()
	var req xlsxExportRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
		legacyError(w, http.StatusBadRequest, "No se pudieron leer los datos para Excel.", nil)
		return
	}
	if len(req.Rows) == 0 || len(req.Rows) > 10000 {
		legacyError(w, http.StatusBadRequest, "Cantidad de filas no valida.", nil)
		return
	}
	inputs := map[int]bool{}
	for _, col := range req.InputColumns {
		if col >= 0 && col < 200 {
			inputs[col] = true
		}
	}
	var out bytes.Buffer
	if err := xlsxrows.Write(&out, req.Rows, xlsxrows.WriteOptions{SheetName: req.Sheet, InputColumns: inputs, Widths: req.Widths, FreezeHeader: true, AutoFilter: true}); err != nil {
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	filename := strings.TrimSpace(req.Filename)
	if filename == "" {
		filename = "Aspirantes_SGL.xlsx"
	}
	filename = strings.ReplaceAll(strings.ReplaceAll(filename, "\r", ""), "\n", "")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(out.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out.Bytes())
}
