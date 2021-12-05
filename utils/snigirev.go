package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

func SnigirevEncrypt(
	key string,
	pin int,
	amount int,
) (nonce string, payload string, err error) {
	keyb := []byte(key)

	var pinw = &bytes.Buffer{}
	err = binary.Write(pinw, binary.LittleEndian, int16(pin))
	if err != nil {
		return nonce, payload, fmt.Errorf("failed to encode pin '%d': %w", pin, err)
	}
	pinb := pinw.Bytes()

	var amountw = &bytes.Buffer{}
	err = binary.Write(amountw, binary.LittleEndian, int32(amount))
	if err != nil {
		return nonce, payload, fmt.Errorf("failed to encode amount '%d': %w", amount, err)
	}
	amountb := amountw.Bytes()

	nonceb := make([]byte, 8)
	rand.Read(nonceb)

	checksum := sha256.Sum256(append(pinb, amountb...))

	h := sha256.New()
	h.Write(nonceb)
	h.Write(keyb)
	s := h.Sum(nil)

	payloadb := s[0:8]
	for i := 0; i < 2; i++ {
		payloadb[i] = payloadb[i] ^ pinb[i]
	}
	for i := 0; i < 4; i++ {
		payloadb[2+i] = payloadb[2+i] ^ amountb[i]
	}
	for i := 0; i < 2; i++ {
		payloadb[6+i] = payloadb[6+i] ^ checksum[i]
	}

	return base64.RawURLEncoding.EncodeToString(nonceb),
		base64.RawURLEncoding.EncodeToString(payloadb),
		nil
}

func SnigirevDecrypt(
	key string,
	nonce string,
	payload string,
) (pin int, amount int, err error) {
	keyb := []byte(key)

	nonceb, err := base64.RawURLEncoding.DecodeString(
		strings.ReplaceAll(nonce, "=", ""))
	if err != nil {
		return pin, amount, fmt.Errorf("nonce '%s' is not valid base64: %w", nonce, err)
	}
	payloadb, err := base64.RawURLEncoding.DecodeString(
		strings.ReplaceAll(payload, "=", ""))
	if err != nil {
		return pin, amount, fmt.Errorf("payload '%s' is not valid base64: %w",
			payload, err)
	}

	// decrypt
	h := sha256.New()
	h.Write(nonceb)
	h.Write(keyb)
	s := h.Sum(nil)
	for i, _ := range payloadb {
		payloadb[i] = payloadb[i] ^ s[i]
	}

	// read integers from decrypted buffer
	var pin16 int16
	err = binary.Read(bytes.NewReader(payloadb[0:2]), binary.LittleEndian, &pin16)
	if err != nil {
		return pin, amount, fmt.Errorf("failed to decrypt pin: %w", err)
	}
	pin = int(pin16)

	var amount32 int32
	err = binary.Read(bytes.NewReader(payloadb[2:6]), binary.LittleEndian, &amount32)
	if err != nil {
		return pin, amount, fmt.Errorf("failed to decrypt amount: %w", err)
	}
	amount = int(amount32)

	// verify checksum (sha256(pin bytes + amount bytes)[0:2] == 2 bytes at the end)
	checksum := sha256.Sum256(payloadb[0:6])
	for i := 0; i < 2; i++ {
		if checksum[i] != payloadb[6+i] {
			return pin, amount, fmt.Errorf("invalid checksum")
		}
	}

	return pin, amount, nil
}
