package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

func SnigirevEncrypt(
	key string,
	pin int,
	amount int,
) (blob string, err error) {
	keyb := []byte(key)

	// generate random nonce
	nonce := make([]byte, 8)
	rand.Read(nonce)

	w := &bytes.Buffer{}
	// encode pin and amount
	if err := WriteVarInt(w, uint64(pin)); err != nil {
		return blob, fmt.Errorf("failed to encode pin: %w", err)
	}
	if err := WriteVarInt(w, uint64(amount)); err != nil {
		return blob, fmt.Errorf("failed to encode amount: %w", err)
	}

	payload := w.Bytes()
	// encrypt payload
	secrethmac := hmac.New(sha256.New, keyb)
	secrethmac.Write([]byte("Round secret:"))
	secrethmac.Write(nonce)
	secret := secrethmac.Sum(nil)
	for i := range payload {
		payload[i] = payload[i] ^ secret[i]
	}

	// generate checksum
	checksumhmac := hmac.New(sha256.New, keyb)
	checksumhmac.Write([]byte("Data:"))
	checksumhmac.Write(nonce)
	checksumhmac.Write(payload)
	checksum := checksumhmac.Sum(nil)

	// concat everything
	output := &bytes.Buffer{}
	output.Write([]byte{1})
	output.Write([]byte{uint8(len(nonce))})
	output.Write(nonce)
	output.Write([]byte{uint8(len(payload))})
	output.Write(payload)
	output.Write(checksum[:])

	return base64.RawURLEncoding.EncodeToString(output.Bytes()), nil
}

func SnigirevDecrypt(key string, b64blob string) (pin int, amount int, err error) {
	keyb := []byte(key)

	blob, err := base64.RawURLEncoding.DecodeString(
		strings.ReplaceAll(b64blob, "=", ""))
	if err != nil {
		return pin, amount,
			fmt.Errorf("blob '%s' is not valid base64url: %w", b64blob, err)
	}

	s := bytes.NewBuffer(blob)

	// extensibility byte that probably will never be used
	variant := make([]byte, 1)
	if _, err := io.ReadFull(s, variant); err != nil {
		return pin, amount, fmt.Errorf("blob is empty: %w", err)
	}
	if variant[0] != 1 {
		return pin, amount,
			fmt.Errorf("encryption scheme %x not implemented", int(variant[0]))
	}

	// read nonce
	l := make([]byte, 1)
	if _, err := io.ReadFull(s, l); err != nil {
		return pin, amount, fmt.Errorf("blob has insufficient bytes on nonce l: %w", err)
	}
	nonce := make([]byte, l[0])
	if _, err := io.ReadFull(s, nonce); err != nil {
		return pin, amount, fmt.Errorf("blob has insufficient bytes on nonce v: %w", err)
	}

	// read payload
	if _, err := io.ReadFull(s, l); err != nil {
		return pin, amount, fmt.Errorf("blob has insufficient bytes on payload l: %w", err)
	}
	payload := make([]byte, l[0])
	if _, err := io.ReadFull(s, payload); err != nil {
		return pin, amount, fmt.Errorf("blob has insufficient bytes on payload v: %w", err)
	}

	// verify checksum
	checksum := s.Bytes()
	if len(checksum) < 8 {
		return pin, amount,
			fmt.Errorf("checksum must be at least 8 bytes, not %d", len(checksum))
	}
	expectedhmac := hmac.New(sha256.New, keyb)
	expectedhmac.Write([]byte("Data:"))
	expectedhmac.Write(blob[:len(blob)-len(checksum)])
	expected := expectedhmac.Sum(nil)
	if !hmac.Equal(checksum, expected[:len(checksum)]) {
		return pin, amount, fmt.Errorf("invalid checksum")
	}

	// decrypt
	secrethmac := hmac.New(sha256.New, keyb)
	secrethmac.Write([]byte("Round secret:"))
	secrethmac.Write(nonce)
	secret := secrethmac.Sum(nil)
	for i := range payload {
		payload[i] = payload[i] ^ secret[i]
	}
	buf := bytes.NewBuffer(payload)

	// read integers from decrypted buffer
	if pin64, err := ReadVarInt(buf); err != nil {
		return pin, amount, fmt.Errorf("failed to read pin from %x: %w", payload, err)
	} else {
		pin = int(pin64)
	}

	if amount64, err := ReadVarInt(buf); err != nil {
		return amount, amount,
			fmt.Errorf("failed to read amount from %x: %w", payload, err)
	} else {
		amount = int(amount64)
	}

	return pin, amount, nil
}
