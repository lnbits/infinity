package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomHex(nbytes int) string {
	b := make([]byte, nbytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}
