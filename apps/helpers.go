package apps

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
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

func urljoin(baseURL *url.URL, elems ...string) *url.URL {
	for _, elem := range elems {
		if strings.HasPrefix(elem, "/") {
			baseURL.Path = elem
		} else if strings.HasPrefix(elem, "http") {
			baseURL, _ = url.Parse(elem)
		} else {
			baseURL.Path = path.Join(baseURL.Path, elem)
		}
	}

	return baseURL
}

func serveFile(w http.ResponseWriter, r *http.Request, fileURL *url.URL) {
	(&httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = fileURL
			r.Host = fileURL.Host
		},
	}).ServeHTTP(w, r)
}
