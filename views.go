package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func viewSettings(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct {
		SiteTitle       string   `json:"siteTitle"`
		SiteTagLine     string   `json:"siteTagline"`
		SiteDescription string   `json:"siteDescription"`
		AllowedThemes   []string `json:"allowedThemes"`
		SiteVersion     string   `json:"siteVersion"`
		Currencies      []string `json:"currencies"`
	}{
		s.SiteTitle,
		s.SiteTagline,
		s.SiteDescription,
		s.ThemeOptions,
		commit,
		CURRENCIES,
	})
}

func viewLnurlScan(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]

	log.Print(code)
}

func viewSSE(w http.ResponseWriter, r *http.Request) {

}
