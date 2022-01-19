package services

import (
	"fmt"
	"net/http"
	"time"
)

var Secret string
var MainServer *http.Server
var TunnelDomain string

var httpClient = &http.Client{
	Timeout: time.Second * 7,
	CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return fmt.Errorf("target '%s' has returned a redirect", r.URL)
	},
}
