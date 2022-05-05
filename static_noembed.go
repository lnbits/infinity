//go:build noembed
// +build noembed

package main

import (
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

func serveStaticClient(router *mux.Router) {
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static",
			http.FileServer(http.Dir("./static/")),
		),
	)

	router.PathPrefix("/").Handler(
		httputil.NewSingleHostReverseProxy(s.QuasarDevServer),
	)
}
