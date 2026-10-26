package main

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

const musicImportFileLimit625 int64 = 240 << 20

type musicImportRequest625 struct {
	Key          string          `json:"key"`
	Password     string          `json:"password"`
	Payload      json.RawMessage `json:"payload"`
	Preview      bool            `json:"preview"`
	Confirmation string          `json:"confirmation"`
}

// The browser submits its File directly as multipart data. It never creates
// a full text string, a parsed object tree and a second JSON string for a large
// backup. The JSON envelope remains supported for existing clients.
func readMusicImport625(w http.ResponseWriter, r *http.Request, limit int64) (musicImportRequest625, error) {
	var req musicImportRequest625
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil && r.Header.Get("Content-Type") != "" {
		return req, errors.New("tipo de archivo de respaldo inválido")
	}
	if kind != "multipart/form-data" {
		if err = readJSONLimited(r, &req, 256<<20); err != nil {
			return req, errors.New("No se pudo leer el respaldo JSON (límite 256 MB).")
		}
		return req, nil
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit+(1<<20))
	reader, err := r.MultipartReader()
	if err != nil {
		return req, errors.New("No se pudo abrir el archivo enviado.")
	}
	seen := map[string]bool{}
	for count := 0; ; count++ {
		part, e := reader.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			return req, errors.New("El envío está incompleto o supera el límite de 240 MB.")
		}
		name := part.FormName()
		if count >= 5 || seen[name] {
			_ = part.Close()
			return req, errors.New("El envío contiene campos duplicados o no válidos.")
		}
		seen[name] = true
		max := int64(8192)
		if name == "backup" {
			max = limit
		}
		b, e := io.ReadAll(io.LimitReader(part, max+1))
		_ = part.Close()
		if e != nil || int64(len(b)) > max {
			return req, errors.New("El archivo supera 240 MB o un campo del formulario es demasiado largo.")
		}
		switch name {
		case "backup":
			if part.FileName() == "" {
				return req, errors.New("Selecciona un archivo de respaldo.")
			}
			req.Payload = b
		case "key":
			req.Key = string(b)
		case "password":
			req.Password = string(b)
		case "preview":
			if string(b) != "true" && string(b) != "false" {
				return req, errors.New("Vista previa inválida.")
			}
			req.Preview = string(b) == "true"
		case "confirmation":
			req.Confirmation = string(b)
		default:
			return req, errors.New("Campo del formulario no reconocido.")
		}
	}
	if len(req.Payload) == 0 || strings.TrimSpace(string(req.Payload[:min(len(req.Payload), 64)])) == "" {
		return req, errors.New("Selecciona un respaldo JSON o .cabackup que contenga datos.")
	}
	return req, nil
}
