package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
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
			if r.URL.Path == "/api/user" || strings.HasPrefix(r.URL.Path, "/api/wallet/") {
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
