package main

import (
	"encoding/json"
	"net/http"

	"github.com/lnbits/lnbits/utils"
)

func viewSettings(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct {
		SiteTitle       string   `json:"siteTitle"`
		SiteTagLine     string   `json:"siteTagline"`
		SiteDescription string   `json:"siteDescription"`
		SiteVersion     string   `json:"siteVersion"`
		Currencies      []string `json:"currencies"`
	}{
		s.SiteTitle,
		s.SiteTagline,
		s.SiteDescription,
		commit,
		utils.CURRENCIES,
	})
}
