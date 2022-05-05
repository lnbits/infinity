//go:build !noembed
// +build !noembed

package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed client/dist/spa
var client embed.FS

//go:embed static
var static embed.FS

func serveStaticClient(router *mux.Router) {
	// serve js library and other assets
	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(static)))

	// serve static client
	if clientFS, err := fs.Sub(client, "client/dist/spa"); err != nil {
		log.Fatal().Err(err).Msg("failed to load static files subdir")
		return
	} else {
		spaFS := SpaFS{clientFS}
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
