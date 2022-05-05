package nostr_utils

import "github.com/fiatjaf/go-nostr"

var (
	pool   = nostr.RelayPool{}
	Secret string
	Relays []string
)
