package apps

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/aarzilli/golua/lua"
)

var (
	float64type = reflect.TypeOf(0.0)
	booltype    = reflect.TypeOf(true)
)

func appidToURL(appid string) string {
	if url, err := base64.StdEncoding.DecodeString(appid); err == nil {
		return string(url)
	} else {
		log.Warn().Err(err).Str("appid", appid).Msg("got invalid app id")
		return ""
	}
}

func stacktrace(luaError *lua.LuaError) string {
	stack := luaError.StackTrace()
	message := []string{luaError.Error() + "\n"}
	for i := len(stack) - 1; i >= 0; i-- {
		entry := stack[i]
		message = append(message, entry.Name+" "+entry.ShortSource)
	}
	return strings.Join(message, "\n")
}

// convert struct to map[string]interface{}
func structToMap(v interface{}) map[string]interface{} {
	j, _ := json.Marshal(v)
	var m map[string]interface{}
	json.Unmarshal(j, &m)
	return m
}
