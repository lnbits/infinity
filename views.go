package main

import (
	"net/http"

	"github.com/lnbits/lnbits/api/apiutils"
	"github.com/lnbits/lnbits/utils"
)

func viewSettings(w http.ResponseWriter, r *http.Request) {
	apiutils.SendJSON(w, struct {
		ServiceURL      string   `json:"serviceURL"`
		SiteTitle       string   `json:"siteTitle"`
		SiteTagLine     string   `json:"siteTagline"`
		SiteDescription string   `json:"siteDescription"`
		SiteVersion     string   `json:"siteVersion"`
		Currencies      []string `json:"currencies"`
	}{
		s.ServiceURL,
		s.SiteTitle,
		s.SiteTagline,
		s.SiteDescription,
		commit,
		utils.CURRENCIES,
	})
}
