// +build !noembed

package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed client/dist/spa
var static embed.FS

func serveStaticClient() {
	// serve static client
	if staticFS, err := fs.Sub(static, "client/dist/spa"); err != nil {
		log.Fatal().Err(err).Msg("failed to load static files subdir")
		return
	} else {
		spaFS := SpaFS{staticFS}
		httpFS := http.FS(spaFS)
		router.PathPrefix("/").Handler(http.FileServer(httpFS))
	}
}

type SpaFS struct {
	base fs.FS
}

func (s SpaFS) Open(name string) (fs.File, error) {
	if file, err := s.base.Open(name); err == nil {
		return file, nil
	} else {
		return s.base.Open("index.html")
	}
}
