package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lnbits/infinity/api/apiutils"
	"github.com/lnbits/infinity/utils"
	"gopkg.in/antage/eventsource.v1"
)

func CustomAction(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appIDToURL(mux.Vars(r)["appid"])
	action := mux.Vars(r)["action"]

	if !nameValidator.MatchString(action) {
		// prevents code injection by public action callers
		apiutils.SendJSONError(w, 400, "invalid action name '%s'", action)
		return
	}

	params := make(map[string]interface{})
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&params)
	}
	querystringFields := make(map[string]struct{})
	for k, v := range r.URL.Query() {
		if _, ok := params[k]; !ok {
			params[k] = v[0]
			querystringFields[k] = struct{}{}
		}
	}

	settings, err := GetAppSettings(app, false)
	if err != nil {
		apiutils.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	def, ok := settings.Actions[action]
	if !ok {
		apiutils.SendJSONError(w, 404, "action '%s' not defined on app", action)
		return
	}

	if err := def.validateParams(params, querystringFields, walletID, app); err != nil {
		apiutils.SendJSONError(w, 400,
			"'%s' called with invalid params: %s", action, err.Error())
		return
	}

	// add special params
	params["_url"] = getOriginalURL(r).String()

	returned, err := runlua(RunluaParams{
		AppURL:          app,
		CodeToRun:       fmt.Sprintf("actions['%s'].handler(internal.arg)", action),
		InjectedGlobals: &map[string]interface{}{"arg": params},
		WalletID:        walletID,
	})
	if err != nil {
		apiutils.SendJSONError(w, 470, "failed to run action: %s", err.Error())
		return
	}

	// if the action returns a table with these values we'll interpret it in a special way
	if complexResponse, ok := returned.(map[string]interface{}); ok {
		ibody, ok1 := complexResponse["body"]
		istatus, ok2 := complexResponse["status"]
		iheaders, ok3 := complexResponse["headers"]
		if ok1 && ok2 && ok3 {
			body, ok1 := ibody.(string)
			status, ok2 := istatus.(float64)
			headers, ok3 := iheaders.(map[string]interface{})
			if !ok1 || !ok2 || !ok3 {
				apiutils.SendJSONError(w, 471,
					"action returned an invalid complex response.")
				return
			}

			w.WriteHeader(int(status))

			for key, ival := range headers {
				if val, ok := ival.(string); ok {
					w.Header().Set(key, val)
				}
			}

			fmt.Fprint(w, body)

			return
		}
	}

	apiutils.SendJSON(w, returned)
}

func PublicSSE(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appIDToURL(mux.Vars(r)["appid"])

	var es eventsource.EventSource
	ies, ok := publicAppStreams.Load(walletID + ":" + app)

	if !ok {
		es = eventsource.New(
			&eventsource.Settings{
				Timeout:        5 * time.Second,
				CloseOnTimeout: true,
				IdleTimeout:    1 * time.Minute,
			},
			func(r *http.Request) [][]byte {
				return [][]byte{
					[]byte("X-Accel-Buffering: no"),
					[]byte("Cache-Control: no-cache"),
					[]byte("Content-Type: text/event-stream"),
					[]byte("Connection: keep-alive"),
					[]byte("Access-Control-Allow-Origin: *"),
				}
			},
		)
		go func() {
			for {
				time.Sleep(25 * time.Second)
				es.SendEventMessage("", "keepalive", "")
			}
		}()

		publicAppStreams.Store(walletID+":"+app, es)
	} else {
		es = ies.(eventsource.EventSource)
	}

	go func() {
		time.Sleep(1 * time.Second)
		es.SendRetryMessage(3 * time.Second)
	}()

	es.ServeHTTP(w, r)
}

func emitPublicEvent(walletID string, app string, typ string, data interface{}) {
	jpayload, _ := utils.JSONMarshal(data)
	payload := string(jpayload)

	if ies, ok := publicAppStreams.Load(walletID + ":" + app); ok {
		ies.(eventsource.EventSource).SendEventMessage(payload, typ, "")
	}
}

func StaticFile(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	subpath := "/" + strings.Join(strings.Split(r.URL.Path, "/")[4:], "/")

	settings, err := GetAppSettings(app, false)
	if err != nil {
		http.Error(w, "failed to get app settings: "+err.Error(), 420)
		return
	}

	baseURL, err := url.Parse(app)
	if err != nil {
		http.Error(w, "failed to get parse app URL: "+err.Error(), 420)
		return
	}

	if specific, ok := settings.Files[subpath]; ok {
		serveFile(w, r, urljoin(*baseURL, specific))
		return
	}

	if catchall, ok := settings.Files["*"]; ok {
		serveFile(w, r, urljoin(*baseURL, catchall))
		return
	}

	serveFile(w, r, urljoin(*baseURL, "index.html"))
}
