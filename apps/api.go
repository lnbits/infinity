package apps

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lnbits/lnbits/api/apiutils"
	"github.com/lnbits/lnbits/models"
)

func Info(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])

	settings, err := GetAppSettings(app)
	if err != nil {
		apiutils.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(settings)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	codeCache.Delete(app)
	settingsCache.Delete(app)
}

func ListItems(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	wallet := r.Context().Value("wallet").(*models.Wallet)
	modelName := mux.Vars(r)["model"]

	settings, err := GetAppSettings(app)
	if err != nil {
		apiutils.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	items, err := DBList(wallet.ID, app, modelName)
	if err != nil {
		apiutils.SendJSONError(w, 500, "database error: %s", err.Error())
		return
	}

	// preprocess items
	/// computed
	model := settings.getModel(modelName)
	for _, field := range model.Fields {
		if field.Computed != nil {
			for _, item := range items {
				var err error
				item.Value[field.Name], err = runlua(RunluaParams{
					AppURL: app,
					CodeToRun: fmt.Sprintf(
						"internal.get_model_field('%s', '%s').computed(internal.arg)",
						model.Name, field.Name,
					),
					InjectedGlobals: &map[string]interface{}{"arg": structToMap(item)},
				})
				if err != nil {
					log.Debug().Err(err).Interface("item", item).
						Str("model", model.Name).Str("field", field.Name).
						Msg("failed to run compute")
				}
			}
		}
	}

	json.NewEncoder(w).Encode(items)
}

func GetItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if value, err := DBGet(wallet.ID, app, model, key); err != nil {
		apiutils.SendJSONError(w, 500, "failed to get item: %s", err.Error())
		return
	} else {
		json.NewEncoder(w).Encode(value)
	}
}

func SetItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
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
}

func AddItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var value map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		apiutils.SendJSONError(w, 400, "failed to read data: %s", err.Error())
		return
	}

	if key, err := DBAdd(wallet.ID, app, model, value); err != nil {
		apiutils.SendJSONError(w, 500, "failed to add item: %s", err.Error())
		return
	} else {
		json.NewEncoder(w).Encode(key)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	model := mux.Vars(r)["model"]
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if err := DBDelete(wallet.ID, app, model, key); err != nil {
		apiutils.SendJSONError(w, 500, "failed to delete item: %s", err.Error())
		return
	}
}
