package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticCompression537(t *testing.T) {
	dir := t.TempDir()
	plain := bytes.Repeat([]byte("const value=123;\n"), 300)
	os.WriteFile(filepath.Join(dir, "app.js"), plain, 0600)
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	gz.Write(plain)
	gz.Close()
	os.WriteFile(filepath.Join(dir, "app.js.gz"), compressed.Bytes(), 0600)
	handler := compressedStatic(dir, http.FileServer(http.Dir(dir)))
	for _, encoding := range []string{"gzip", "br, gzip;q=0.8", "gzip;q=0", ""} {
		req := httptest.NewRequest("GET", "/app.js", nil)
		req.Header.Set("Accept-Encoding", encoding)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		body := res.Body.Bytes()
		if encoding == "gzip" || encoding == "br, gzip;q=0.8" {
			if res.Header().Get("Content-Encoding") != "gzip" {
				t.Fatal("missing gzip")
			}
			r, e := gzip.NewReader(bytes.NewReader(body))
			if e != nil {
				t.Fatal(e)
			}
			body, _ = io.ReadAll(r)
		} else if res.Header().Get("Content-Encoding") != "" {
			t.Fatal("ignored gzip q=0")
		}
		if !bytes.Equal(body, plain) {
			t.Fatal("content changed")
		}
	}
	req := httptest.NewRequest("GET", "/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Range", "bytes=0-4")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != 206 || res.Body.String() != string(plain[:5]) {
		t.Fatal("range changed")
	}
}
