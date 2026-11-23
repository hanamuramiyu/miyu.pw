package webui

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

const (
	contentSecurityPolicy = "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'"
	indexCache            = "no-cache"
	assetCache            = "public, max-age=31536000, immutable"
)

func Handler() http.Handler {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		return New(nil)
	}
	return New(dist)
}

func New(fsys fs.FS) http.Handler {
	return &handler{fsys: fsys}
}

type handler struct {
	fsys fs.FS
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.built() {
		servePlaceholder(w)
		return
	}

	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" {
		h.serveIndex(w, r)
		return
	}

	if strings.HasPrefix(name, "assets/") {
		if !h.serveFile(w, r, name, assetCache) {
			http.NotFound(w, r)
		}
		return
	}

	if h.serveFile(w, r, name, indexCache) {
		return
	}

	h.serveIndex(w, r)
}

func (h *handler) built() bool {
	if h.fsys == nil {
		return false
	}
	info, err := fs.Stat(h.fsys, "index.html")
	return err == nil && !info.IsDir()
}

func (h *handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	if !h.serveFile(w, r, "index.html", indexCache) {
		servePlaceholder(w)
	}
}

func (h *handler) serveFile(w http.ResponseWriter, r *http.Request, name, cache string) bool {
	file, err := h.fsys.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		return false
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return false
	}

	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", cache)
	http.ServeContent(w, r, name, info.ModTime(), bytes.NewReader(data))
	return true
}

func setHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Robots-Tag", "noindex")
	header.Set("Content-Security-Policy", contentSecurityPolicy)
}

func servePlaceholder(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, placeholderHTML)
}

const placeholderHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex">
<title>miyu admin - frontend not built</title>
<style>
body { margin: 0; padding: 3rem 1.5rem; background: #141211; color: #ebe4dc; font: 16px/1.5 system-ui, sans-serif; }
main { max-width: 42rem; margin: 0 auto; }
h1 { margin: 0 0 1rem; font-size: 1.75rem; font-weight: 600; }
p { margin: 0 0 1rem; color: #a89d94; }
pre { margin: 1.5rem 0 0; padding: 1rem; overflow-x: auto; border: 1px solid #2c2825; color: #ebe4dc; }
</style>
</head>
<body>
<main>
<h1>frontend not built</h1>
<p>The admin frontend is not compiled into this binary. Build it, then rebuild the server.</p>
<pre>cd admin/web
bun install --frozen-lockfile
bun run build
cd ..
go build ./cmd/server</pre>
</main>
</body>
</html>
`
