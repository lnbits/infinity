package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"os"

	"github.com/btcsuite/btcd/btcec"
)

func AuthKey(walletID string, domain string) *btcec.PrivateKey {
	hashingKey := sha256.Sum256([]byte(os.Getenv("SECRET") + walletID))

	h := hmac.New(sha256.New, hashingKey[:])
	h.Write([]byte(domain))

	linkingKey := h.Sum(nil)

	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), linkingKey)

	return priv
}
