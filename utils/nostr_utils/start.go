package nostr_utils

import (
	"crypto/sha256"
	"fmt"

	"github.com/fiatjaf/bip340"
	nostr "github.com/fiatjaf/go-nostr"
	"github.com/lnbits/infinity/events"
)

func Start() {
	if len(Relays) == 0 {
		return
	}

	privateKeyBytes := sha256.Sum256([]byte(Secret + ":nostrkey"))
	privateKeyHex := fmt.Sprintf("%x", privateKeyBytes[:])
	privateKeyN, _ := bip340.ParsePrivateKey(privateKeyHex)
	publicKey := bip340.GetPublicKey(privateKeyN)

	pool.SecretKey = &privateKeyHex

	for _, url := range Relays {
		pool.Add(url, &nostr.SimplePolicy{Read: true, Write: true})
	}

	sub := pool.Sub(nostr.Filters{
		nostr.Filter{
			Tags: map[string]nostr.StringList{
				"#p": {fmt.Sprintf("%x", publicKey)},
			},
		},
	})

	go func() {
		for event := range sub.UniqueEvents {
			events.EmitGenericEvent("nostr_event", event)
		}
	}()
}
