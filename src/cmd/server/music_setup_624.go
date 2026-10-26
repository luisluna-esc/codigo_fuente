package main

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Only a direct loopback visit may fill the one-use installation key. Remote
// and reverse-proxy visits still require the key from the server's file.
func (s *Server) localSetupKey624(r *http.Request) string {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || net.ParseIP(peer) == nil || !net.ParseIP(peer).IsLoopback() {
		return ""
	}
	host := r.Host
	if name, _, e := net.SplitHostPort(host); e == nil {
		host = name
	}
	if host != "127.0.0.1" && host != "localhost" && host != "::1" && host != "[::1]" {
		return ""
	}
	for _, header := range []string{"Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Real-IP"} {
		if r.Header.Get(header) != "" {
			return ""
		}
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, e := url.Parse(origin)
		if e != nil || !strings.EqualFold(u.Host, r.Host) || u.Scheme != "http" && u.Scheme != "https" {
			return ""
		}
	}
	data, err := os.ReadFile(s.installationKeyPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
