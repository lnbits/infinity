// +build noembed

package main

import "net/http/httputil"

func serveStaticClient() {
	router.PathPrefix("/").Handler(
		httputil.NewSingleHostReverseProxy(s.QuasarDevServer),
	)
}
