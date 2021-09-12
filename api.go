package main

import (
	"encoding/json"
	"net/http"
)

func apiSettings(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct {
		SiteTitle     string   `json:"siteTitle"`
		SiteTagLine   string   `json:"siteTagline"`
		AllowedThemes []string `json:"allowedThemes"`
		SiteVersion   string   `json:"siteVersion"`
	}{
		s.SiteTitle,
		s.SiteTagline,
		s.ThemeOptions,
		commit,
	})
}

func apiUser(w http.ResponseWriter, r *http.Request) {

}

func apiWallet(w http.ResponseWriter, r *http.Request) {

}
