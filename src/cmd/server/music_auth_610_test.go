package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTrip610 func(*http.Request) (*http.Response, error)

func (f roundTrip610) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func Test610GoogleSignedIdentityValidation(t *testing.T) {
	t.Setenv("MUSICA_GOOGLE_CLIENT_ID", "client610")
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	b64 := func(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
	jwks, _ := json.Marshal(map[string]any{"keys": []any{map[string]any{"kty": "RSA", "kid": "test610", "alg": "RS256", "n": b64(private.N.Bytes()), "e": b64([]byte{1, 0, 1})}}})
	previous := identityHTTP610
	identityHTTP610 = &http.Client{Transport: roundTrip610(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://www.googleapis.com/oauth2/v3/certs" {
			t.Fatal("unexpected request", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(jwks)))}, nil
	})}
	defer func() { identityHTTP610 = previous }()
	sign := func(claims map[string]any) string {
		h, _ := json.Marshal(map[string]any{"alg": "RS256", "kid": "test610"})
		c, _ := json.Marshal(claims)
		input := b64(h) + "." + b64(c)
		sum := sha256.Sum256([]byte(input))
		sig, e := rsa.SignPKCS1v15(rand.Reader, private, crypto.SHA256, sum[:])
		if e != nil {
			t.Fatal(e)
		}
		return input + "." + b64(sig)
	}
	claims := map[string]any{"iss": "https://accounts.google.com", "aud": "client610", "nonce": "one-login", "sub": "subject610", "email": "person@gmail.com", "email_verified": true, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Minute).Unix()}
	if got, e := googleClaims610(sign(claims), "one-login"); e != nil || got["sub"] != "subject610" {
		t.Fatal("valid identity", e)
	}
	for _, tc := range []struct {
		name, key string
		value     any
	}{{"audience", "aud", "another-client"}, {"issuer", "iss", "https://untrusted.test"}, {"nonce", "nonce", "another-login"}, {"expired", "exp", time.Now().Add(-time.Minute).Unix()}, {"unverified-email", "email_verified", false}, {"missing-subject", "sub", ""}} {
		t.Run(tc.name, func(t *testing.T) {
			copy := cloneMap610(claims)
			copy[tc.key] = tc.value
			if _, e := googleClaims610(sign(copy), "one-login"); e == nil {
				t.Fatal("invalid identity accepted")
			}
		})
	}
	token := sign(claims)
	parts := strings.Split(token, ".")
	parts[2] = b64(make([]byte, 256))
	if _, e := googleClaims610(strings.Join(parts, "."), "one-login"); e == nil {
		t.Fatal("forged signature accepted")
	}
}
