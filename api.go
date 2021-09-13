package main

import (
	"encoding/json"
	"net/http"
)

func apiSettings(w http.ResponseWriter, r *http.Request) {
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

func apiUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	json.NewEncoder(w).Encode(user)
}

func apiCreateWallet(w http.ResponseWriter, r *http.Request) {
	var masterKey string
	user := &User{}

	if r.Context().Value("user") != nil {
		user = r.Context().Value("user").(*User)
	} else {
		// create user
		user.Apps = make(StringList, 0)
		masterKey := randomHex(32) // will only be returned if we're creating the user
		user.MasterKey = masterKey
		db.Create(user)
	}

	// create wallet
	var params struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		jsonError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	wallet := Wallet{
		Name:       params.Name,
		UserID:     user.ID,
		InvoiceKey: randomHex(32),
		AdminKey:   randomHex(32),
	}
	result := db.Create(&wallet)
	if result.Error != nil {
		jsonError(w, 400, "error saving wallet: %s", result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(struct {
		UserMasterKey string `json:"userMasterKey"`
		Wallet        Wallet `json:"wallet"`
	}{
		masterKey,
		wallet,
	})
}

func apiWallet(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	json.NewEncoder(w).Encode(user)
}

func apiCreateInvoice(w http.ResponseWriter, r *http.Request) {}

func apiPayInvoice(w http.ResponseWriter, r *http.Request) {}

func apiLnurlScan(w http.ResponseWriter, r *http.Request) {}

func apiLnurlAuth(w http.ResponseWriter, r *http.Request) {}

func apiPayLnurl(w http.ResponseWriter, r *http.Request) {}
