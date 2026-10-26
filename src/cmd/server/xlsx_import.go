package main

import (
	"net/http"

	"sgl.local/sistema-gestion-liderazgo/internal/xlsxrows"
)

func (s *Server) importXlsx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, http.StatusForbidden, "Esta ruta requiere Administrador.", nil)
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		legacyError(w, http.StatusBadRequest, "No se pudo leer el archivo Excel.", nil)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		legacyError(w, http.StatusBadRequest, "Falta el archivo Excel.", nil)
		return
	}
	defer file.Close()
	rows, err := xlsxrows.Read(file, 12<<20)
	if err != nil {
		legacyError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	jsonOut(w, http.StatusOK, map[string]any{"ok": true, "rows": rows})
}
