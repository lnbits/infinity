package utils

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/fiatjaf/go-lnurl"
)

// a string/base64 only interface for aes, intended to be used by lua apps

func AESEncrypt(plainText string, b64Key string) (cipherText string, err error) {
	key, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key '%s': %w", b64Key, err)
	}

	ct, iv, err := lnurl.AESCipher(key, []byte(plainText))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	return (base64.StdEncoding.EncodeToString(ct) +
		"?" +
		base64.StdEncoding.EncodeToString(iv)), nil
}

func AESDecrypt(cipherTextPlusIV string, b64Key string) (plainText string, err error) {
	key, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key '%s': %w", b64Key, err)
	}

	spl := strings.Split(cipherTextPlusIV, "?")
	if len(spl) != 2 {
		return "", fmt.Errorf("ciphertext should be in format 'ciphertext?iv'")
	}

	b64CipherText := spl[0]
	b64IV := spl[1]

	cipherText, err := base64.StdEncoding.DecodeString(b64CipherText)
	if err != nil {
		return "", fmt.Errorf("invalid base64 ciphertext: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(b64IV)
	if err != nil {
		return "", fmt.Errorf("invalid base64 iv: %w", err)
	}

	plainTextBytes, err := lnurl.AESDecipher(key, cipherText, iv)
	if err != nil {
		return "", err
	}

	return string(plainTextBytes), nil
}
