package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lnbits/lnbits/api/apiutils"
	"gopkg.in/antage/eventsource.v1"
)

func CustomAction(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appidToURL(mux.Vars(r)["appid"])
	action := mux.Vars(r)["action"]

	if !nameValidator.MatchString(action) {
		// prevents code injection by public action callers
		apiutils.SendJSONError(w, 400, "invalid action name '%s'", action)
		return
	}

	var params map[string]interface{}

	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&params)
	}

	settings, err := GetAppSettings(app)
	if err != nil {
		apiutils.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	def, ok := settings.Actions[action]
	if !ok {
		apiutils.SendJSONError(w, 404, "action '%s' not defined on app", action)
		return
	}

	if err := def.validateParams(params); err != nil {
		apiutils.SendJSONError(w, 400,
			"'%s' called with invalid params: %s", action, err.Error())
		return
	}

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

	json.NewEncoder(w).Encode(returned)
}

func PublicSSE(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appidToURL(mux.Vars(r)["appid"])

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
	jpayload, _ := json.Marshal(data)
	payload := string(jpayload)

	if ies, ok := publicAppStreams.Load(walletID + ":" + app); ok {
		ies.(eventsource.EventSource).SendEventMessage(payload, typ, "")
	}
}

func StaticFile(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	subpath := "/" + strings.Join(strings.Split(r.URL.Path, "/")[4:], "/")

	settings, err := GetAppSettings(app)
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

	serveFile(w, r, urljoin(*baseURL, "/index.html"))
}
