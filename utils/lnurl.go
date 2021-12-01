package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/fiatjaf/go-lnurl"
)

func AESSuccessAction(plaintext string, preimage string) (string, string, error) {
	key, err := hex.DecodeString(preimage)
	if err != nil {
		return "", "", fmt.Errorf("invalid hex preimage '%s': %w", preimage, err)
	}

	ct, iv, err := lnurl.AESCipher(key, []byte(plaintext))
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ct), base64.StdEncoding.EncodeToString(iv), nil
}
