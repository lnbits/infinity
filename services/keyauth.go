package services

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcec/v2"
)

func AuthKey(walletID string, domain string) *btcec.PrivateKey {
	hashingKey := sha256.Sum256([]byte(Secret + walletID))

	h := hmac.New(sha256.New, hashingKey[:])
	h.Write([]byte(domain))

	linkingKey := h.Sum(nil)

	priv, _ := btcec.PrivKeyFromBytes(linkingKey)

	return priv
}
