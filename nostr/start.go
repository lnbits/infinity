package nostr

import (
	"crypto/sha256"
	"fmt"

	"github.com/fiatjaf/bip340"
	"github.com/fiatjaf/go-nostr/filter"
	"github.com/fiatjaf/go-nostr/relaypool"
	"github.com/lnbits/lnbits/events"
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
		pool.Add(url, &relaypool.Policy{
			SimplePolicy: relaypool.SimplePolicy{Read: true, Write: true},
		})
	}

	sub := pool.Sub(filter.EventFilters{
		{
			TagProfile: fmt.Sprintf("%x", publicKey),
		},
	})

	go func() {
		for event := range sub.UniqueEvents {
			events.EmitGenericEvent("nostr_event", event)
		}
	}()
}
