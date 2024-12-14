package gui

import (
	"embed"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	//go:embed static/*
	staticFS embed.FS

	//go:embed tmpl/base.tmpl
	baseFS embed.FS

	//go:embed tmpl/view/*
	viewFS embed.FS

	base *template.Template
)

func init() {
	base = loadBaseTemplate()
}

func loadBaseTemplate() *template.Template {
	return template.Must(template.ParseFS(baseFS, "tmpl/*.tmpl"))
}

func StaticFileServer() http.Handler {
	return NoDirFileServer(http.FS(staticFS))
}

func Base() *template.Template {
	if base == nil {
		base = loadBaseTemplate()
	}

	return base
}

func ViewFS(fileName string) (*template.Template, error) {
	t, err := template.ParseFS(viewFS, filepath.Join("tmpl", "view", fileName))
	if err != nil {
		return nil, err
	}
	return t.Lookup("view.html"), nil
}
