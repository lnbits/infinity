package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lnbits/lnbits/api/apiutils"
)

func CustomAction(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appidToURL(mux.Vars(r)["appid"])
	action := mux.Vars(r)["action"]

	var params map[string]interface{}
	json.NewDecoder(r.Body).Decode(&params)

	settings, err := GetAppSettings(app)
	if err != nil {
		apiutils.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	if _, ok := settings.Actions[action]; !ok {
		apiutils.SendJSONError(w, 404, "action '%s' not defined on app: %s", action, err.Error())
		return
	}

	returned, err := runlua(RunluaParams{
		AppURL:          app,
		CodeToRun:       fmt.Sprintf("actions.%s(internal.arg)", action),
		InjectedGlobals: &map[string]interface{}{"arg": params},
		WalletID:        walletID,
	})
	if err != nil {
		apiutils.SendJSONError(w, 470, "failed to run action: %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(returned)
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
		serveFile(w, r, urljoin(baseURL, specific))
		return
	}

	if catchall, ok := settings.Files["*"]; ok {
		serveFile(w, r, urljoin(baseURL, catchall))
		return
	}

	serveFile(w, r, urljoin(baseURL, "/index.html"))
}
