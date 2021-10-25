package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lnbits/lnbits/api/apiutils"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func jsonHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/v/") {
			w.Header().Set("Content-Type", "application/json")
		}

		next.ServeHTTP(w, r)
	})
}

func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/user") {
			next.ServeHTTP(w, r)
			return
		}

		var user models.User
		var err error
		masterKey := r.Header.Get("X-MasterKey")
		if masterKey == "" {
			err = fmt.Errorf("X-MasterKey header not provided")
		} else {
			err = storage.DB.Where("master_key", masterKey).First(&user).Error
		}

		if err != nil {
			// the user is required for /api/user, but not for /api/create-wallet
			if r.URL.Path != "/api/user/create-wallet" {
				apiutils.SendJSONError(w, 401, "error fetching user: %s", err.Error())
				return
			}
		} else {
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					"user",
					&user,
				),
			)
		}

		next.ServeHTTP(w, r)
	})
}

func walletMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/wallet") && // better API routes
			!strings.HasPrefix(r.URL.Path, "/api/v1/") { // lnbits-compatibility
			next.ServeHTTP(w, r)
			return
		}

		var permission string
		var wallet models.Wallet
		var err error

		// try header
		walletKey := r.Header.Get("X-Api-Key")
		if walletKey == "" {
			// try querystring
			walletKey = r.URL.Query().Get("api-key")
		}

		if walletKey == "" {
			err = fmt.Errorf("X-Api-Key header not provided")
		} else {
			result := storage.DB.Where("admin_key", walletKey).First(&wallet)
			permission = "admin"
			if wallet.ID == "" {
				result = storage.DB.Where("invoice_key", walletKey).First(&wallet)
				permission = "invoice"
				if wallet.ID == "" {
					err = result.Error
				}
			}
		}

		if err != nil {
			apiutils.SendJSONError(w, 401, "error fetching wallet: %s", err.Error())
			return
		} else {
			r = r.WithContext(
				context.WithValue(
					context.WithValue(
						r.Context(),
						"wallet",
						&wallet,
					),
					"permission",
					permission,
				),
			)
		}

		next.ServeHTTP(w, r)
	})
}
