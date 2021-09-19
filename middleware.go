package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user" && r.URL.Path != "/api/create-wallet" {
			next.ServeHTTP(w, r)
			return
		}

		var user User
		var err error
		masterKey := r.Header.Get("X-MasterKey")
		if masterKey == "" {
			err = fmt.Errorf("X-MasterKey header not provided")
		} else {
			err = db.Where("master_key", masterKey).First(&user).Error
		}

		if err != nil {
			// the user is required for /api/user, but not for /api/create-wallet
			if r.URL.Path != "/api/create-wallet" {
				jsonError(w, 500, "error fetching user: %s", err.Error())
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
		if !strings.HasPrefix(r.URL.Path, "/api/v1/") && // lnbits-compatibility
			!strings.HasPrefix(r.URL.Path, "/api/wallet/") { // better API routes
			next.ServeHTTP(w, r)
			return
		}

		var permission string
		var wallet Wallet
		var err error
		walletKey := r.Header.Get("X-Api-Key")
		if walletKey == "" {
			err = fmt.Errorf("X-Api-Key header not provided")
		} else {
			result := db.Where("invoice_key", walletKey).First(&wallet)
			permission = "invoice"
			if wallet.ID == "" {
				result = db.Where("admin_key", walletKey).First(&wallet)
				permission = "admin"
				if wallet.ID == "" {
					err = result.Error
				}
			}
		}

		if err != nil {
			jsonError(w, 500, "error fetching wallet: %s", err.Error())
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
