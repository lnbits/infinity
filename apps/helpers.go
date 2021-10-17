package apps

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	float64type = reflect.TypeOf(0.0)
	booltype    = reflect.TypeOf(true)
)

func appidToURL(appid string) string {
	if url, err := hex.DecodeString(appid); err == nil {
		return string(url)
	} else {
		log.Warn().Err(err).Str("appid", appid).Msg("got invalid app id")
		return ""
	}
}

var reNumber = regexp.MustCompile("\\d+")

func stackTraceWithCode(stacktrace string, code string) string {
	var result []string

	stlines := strings.Split(stacktrace, "\n")
	lines := strings.Split(code, "\n")

	for i := 0; i < len(stlines); i++ {
		stline := stlines[i]
		result = append(result, stline)

		snum := reNumber.FindString(stline)
		if snum != "" {
			num, _ := strconv.Atoi(snum)
			for i, line := range lines {
				line = fmt.Sprintf("%3d %s", i+1, line)
				if i+1 > num-3 && i+1 < num+3 {
					result = append(result, line)
				}
			}
		}
	}

	return strings.Join(result, "\n")
}
