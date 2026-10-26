package main

import (
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Serve precompressed text assets without spending CPU on each request.
// Dynamic/API responses and range requests retain their original behavior.
func compressedStatic(root string, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + r.URL.Path)
		if clean == "/" {
			clean = "/index.html"
		}
		ext := strings.ToLower(filepath.Ext(clean))
		switch ext {
		case ".js", ".css", ".html", ".json", ".webmanifest", ".svg":
		default:
			fallback.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		accepts := false
		for _, entry := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
			bits := strings.Split(strings.TrimSpace(entry), ";")
			if !strings.EqualFold(strings.TrimSpace(bits[0]), "gzip") {
				continue
			}
			quality := 1.0
			for _, param := range bits[1:] {
				if pair := strings.SplitN(strings.TrimSpace(param), "=", 2); len(pair) == 2 && strings.EqualFold(pair[0], "q") {
					value, err := strconv.ParseFloat(pair[1], 64)
					if err != nil {
						quality = 0
					} else {
						quality = value
					}
				}
			}
			accepts = quality > 0
		}
		if !accepts || r.Header.Get("Range") != "" || (r.Method != "GET" && r.Method != "HEAD") {
			fallback.ServeHTTP(w, r)
			return
		}
		filename := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
		original, err := os.Stat(filename)
		if err != nil || original.IsDir() {
			fallback.ServeHTTP(w, r)
			return
		}
		f, err := os.Open(filename + ".gz")
		if err != nil {
			fallback.ServeHTTP(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil || info.ModTime().Before(original.ModTime()) {
			fallback.ServeHTTP(w, r)
			return
		}
		if contentType := mime.TypeByExtension(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(w, r, path.Base(clean), original.ModTime(), f)
	})
}
