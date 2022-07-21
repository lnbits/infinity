package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/fiatjaf/go-lnurl"
)

func PerformKeyAuthFlow(key *btcec.PrivateKey, callback *url.URL, k1 []byte) error {
	qs := callback.Query()
	qs.Set("k1", hex.EncodeToString(k1))
	qs.Set("key", hex.EncodeToString(key.PubKey().SerializeCompressed()))
	qs.Set("sig", hex.EncodeToString(ecdsa.Sign(key, k1).Serialize()))

	callback.RawQuery = qs.Encode()
	targetURL := callback.String()

	resp, err := httpClient.Get(targetURL)
	if err != nil {
		return fmt.Errorf("error in http call: %w", err)
	}
	defer resp.Body.Close()

	var reply lnurl.LNURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return fmt.Errorf("invalid JSON response from %s: %w", callback.Host, err)
	}

	if reply.Status == "ERROR" {
		return fmt.Errorf(reply.Reason)
	}

	return nil
}
