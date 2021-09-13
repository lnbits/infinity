package main

import (
	"context"
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
		err := db.Where("master_key", r.Header.Get("X-MasterKey")).First(&user).Error
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
