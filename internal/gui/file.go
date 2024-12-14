package gui

import (
	"fmt"
	"net/http"
	"syscall"
)

// NoDirStaticHandler returns an http.Handler
// The function assumes that directory path where the static files reside on disk
// are the same to url path prefix
func NoDirStaticHandler(prefix string) http.Handler {
	dir := http.Dir(fmt.Sprintf(".%s", prefix))
	fs := http.FileServer(NoDirFileSystem{dir})
	return http.StripPrefix(prefix, fs)
}

func NoDirFileServer(fs http.FileSystem) http.Handler {
	return http.FileServer(NewNoDirFileSystem(fs))
}

type NoDirFileSystem struct {
	fs http.FileSystem
}

func (fs NoDirFileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		return nil, syscall.ENOENT
	}

	return f, nil
}

func NewNoDirFileSystem(fs http.FileSystem) NoDirFileSystem {
	return NoDirFileSystem{fs: fs}
}
