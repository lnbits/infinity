// +build noembed

package main

import (
	"net/http"
	"net/http/httputil"
)

func serveStaticClient() {
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static",
			http.FileServer(http.Dir("./static/")),
		),
	)

	router.PathPrefix("/").Handler(
		httputil.NewSingleHostReverseProxy(s.QuasarDevServer),
	)
}
