package apps

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/aarzilli/golua/lua"
	"github.com/lnbits/lnbits/utils"
)

var (
	float64type = reflect.TypeOf(0.0)
	booltype    = reflect.TypeOf(true)

	staticresourcetransport = &staticResourceTransport{}
)

func appidToURL(appid string) string {
	appid = strings.ReplaceAll(appid, "=", "")
	if url, err := base64.RawURLEncoding.DecodeString(appid); err == nil {
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
	j, _ := utils.JSONMarshal(v)
	var m map[string]interface{}
	json.Unmarshal(j, &m)
	return m
}

func urljoin(baseURL url.URL, elems ...string) *url.URL {
	for _, elem := range elems {
		if strings.HasPrefix(elem, "/") {
			baseURL.Path = elem
		} else if strings.HasPrefix(elem, "http") {
			newBaseURL, _ := url.Parse(elem)
			return newBaseURL
		} else if strings.HasSuffix(baseURL.Path, "/") {
			baseURL.Path = path.Join(baseURL.Path, elem)
		} else {
			spl := strings.Split(baseURL.Path, "/")
			pathWithoutLastPart := strings.Join(spl[0:len(spl)-1], "/")
			baseURL.Path = path.Join(pathWithoutLastPart, elem)
		}
	}

	return &baseURL
}

func serveFile(w http.ResponseWriter, r *http.Request, fileURL *url.URL) {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = fileURL
			r.Host = fileURL.Host
		},
		Transport: staticresourcetransport,
		ModifyResponse: func(w *http.Response) error {
			if w.StatusCode >= 400 {
				response, _ := ioutil.ReadAll(w.Body)
				w.Body.Close()
				w.Body = io.NopCloser(bytes.NewBuffer(
					[]byte(fmt.Sprintf("%s said: %s", fileURL, string(response))),
				))
			}
			return nil
		},
	}

	proxy.ServeHTTP(w, r)
}

// this is an http.RoundTripper that only uses the URL and ignores the rest of the request
// used only to proxy the static files for the apps on serveFile above.
// this is needed because the default transport was causing Caddy to return a 400
// for mysterious reasons that would be too painful and useless to investigate.
type staticResourceTransport struct{}

func (_ *staticResourceTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return httpClient.Get(r.URL.String())
}

func getOriginalURL(r *http.Request) *url.URL {
	if ServiceURL != "" {
		if serviceURL, err := url.Parse(ServiceURL); err == nil {
			return serviceURL
		} else {
			log.Warn().Err(err).Str("ServiceURL", ServiceURL).
				Msg("SERVICE_URL is invalid")
		}
	}

	if r.URL.Host == "" {
		r.URL.Host = r.Header.Get("X-Forwarded-Host")
		if r.URL.Host == "" {
			for _, entry := range strings.Split(r.Header.Get("Forwarded"), ";") {
				spl := strings.Split(strings.TrimSpace(entry), "=")
				if spl[0] == "host" {
					r.URL.Host = spl[1]
					break
				}
			}
		}
	}
	if r.URL.Scheme == "" {
		r.URL.Scheme = r.Header.Get("X-Forwarded-Proto")
		if r.URL.Scheme == "" {
			r.URL.Scheme = r.Header.Get("X-Forwarded-Protocol")
			if r.URL.Scheme == "" {
				r.URL.Scheme = r.Header.Get("X-Url-Scheme")
				if r.URL.Scheme == "" && r.Header.Get("X-Forwarded-Ssl") == "on" {
					r.URL.Scheme = "https"
				}
				if r.URL.Scheme == "" {
					for _, entry := range strings.Split(r.Header.Get("Forwarded"), ";") {
						spl := strings.Split(strings.TrimSpace(entry), "=")
						if spl[0] == "proto" {
							r.URL.Scheme = spl[1]
							break
						}
					}
				}
				if r.URL.Scheme == "" {
					r.URL.Scheme = "https"
				}
			}
		}
	}

	return r.URL
}
