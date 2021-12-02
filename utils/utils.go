package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 7,
	CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return fmt.Errorf("target '%s' has returned a redirect", r.URL)
	},
}

func RandomHex(nbytes int) string {
	b := make([]byte, nbytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Sha256String(preimage string) string {
	hash := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(hash[:])
}
