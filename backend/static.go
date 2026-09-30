package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// на сервере бэк сам отдаёт собранный фронт: /api — запросы, всё остальное — файлы из dist.
// Неизвестные пути отдают index.html, чтобы приложение открывалось с любой ссылки
func withStatic(api http.Handler, dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		path := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
}
