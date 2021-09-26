package apps

import (
	"encoding/hex"

	"github.com/rs/zerolog/log"
)

func appidToURL(appid string) string {
	if url, err := hex.DecodeString(appid); err == nil {
		return string(url)
	} else {
		log.Warn().Err(err).Str("appid", appid).Msg("got invalid app id")
		return ""
	}
}
