package apps

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lnbits/lnbits/api/apiutils"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func Info(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])

	settings, err := GetAppSettings(app)
	if err != nil {
		apiutils.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	apiutils.SendJSON(w, settings)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	codeCache.Delete(app)
	settingsCache.Delete(app)
}

func ClearData(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	wallet := r.Context().Value("wallet").(*models.Wallet)

	result := storage.DB.
		Where(&models.AppDataItem{
			App:      app,
			WalletID: wallet.ID,
		}).
		Delete(&models.AppDataItem{})

	if result.Error != nil {
		apiutils.SendJSONError(w, 500, "database error: %s", result.Error.Error())
		return
	}
}

func ListItems(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	app := appIDToURL(mux.Vars(r)["appid"])
	wallet := r.Context().Value("wallet").(*models.Wallet)
	modelName := mux.Vars(r)["model"]

	items, err := DBList(wallet.ID, app, modelName, qs.Get("startkey"), qs.Get("endkey"))
	if err != nil {
		apiutils.SendJSONError(w, 500, "database error: %s", err.Error())
		return
	}

	apiutils.SendJSON(w, items)
}

func GetItem(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if value, err := DBGet(wallet.ID, app, model, key); err != nil {
		apiutils.SendJSONError(w, 500, "failed to get item: %s", err.Error())
		return
	} else {
		apiutils.SendJSON(w, value)
	}
}

func SetItem(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var value map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		apiutils.SendJSONError(w, 400, "failed to read data: %s", err.Error())
		return
	}

	if err := DBSet(wallet.ID, app, model, key, value); err != nil {
		apiutils.SendJSONError(w, 500, "failed to set item: %s", err.Error())
		return
	}

	go TriggerEventOnSpecificAppWallet(
		AppWallet{wallet.ID, app},
		"api_db_set",
		KeyValue{key, value},
	)
}

func AddItem(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var value map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		apiutils.SendJSONError(w, 400, "failed to read data: %s", err.Error())
		return
	}

	key, err := DBAdd(wallet.ID, app, model, value)
	if err != nil {
		apiutils.SendJSONError(w, 500, "failed to add item: %s", err.Error())
		return
	}

	apiutils.SendJSON(w, key)
	go TriggerEventOnSpecificAppWallet(
		AppWallet{wallet.ID, app},
		"api_db_set", // because set and add are the same, let's simplify
		KeyValue{key, value},
	)
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	app := appIDToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if err := DBDelete(wallet.ID, app, model, key); err != nil {
		apiutils.SendJSONError(w, 500, "failed to delete item: %s", err.Error())
		return
	}

	go TriggerEventOnSpecificAppWallet(
		AppWallet{wallet.ID, app},
		"api_db_delete",
		KeyValue{key, nil},
	)
}
