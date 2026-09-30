package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWithStatic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<div id=root></div>"), 0o600)
	os.MkdirAll(filepath.Join(dir, "assets"), 0o700)
	os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0o600)
	secret := filepath.Join(filepath.Dir(dir), "secret.txt")
	os.WriteFile(secret, []byte("секрет"), 0o600)

	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("api")) })
	h := withStatic(api, dir)

	cases := map[string]string{
		"/api/health":              "api",
		"/":                        "<div id=root></div>",
		"/assets/app.js":           "console.log(1)",
		"/documents":               "<div id=root></div>",
		"/../secret.txt":           "",
		"/assets/../../secret.txt": "",
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "http://x"+strings.Replace(path, "..", "%2e%2e", -1), nil))
		if !strings.Contains(rec.Body.String(), want) || strings.Contains(rec.Body.String(), "секрет") {
			t.Errorf("%s: %q", path, rec.Body.String())
		}
	}
}
