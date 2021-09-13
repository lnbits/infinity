package main

import (
	"crypto/rand"
	"encoding/hex"
)

func randomHex(nbytes int) string {
	b := make([]byte, nbytes)
	rand.Read(b)
	return hex.EncodeToString(b)
}
