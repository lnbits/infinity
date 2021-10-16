package apps

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	api "github.com/lnbits/lnbits/api"
	"github.com/lnbits/lnbits/apps/db"
	"github.com/lnbits/lnbits/apps/runlua"
	models "github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func AppInfo(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	_, settings, err := getAppSettings(app)
	if err != nil {
		api.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(settings)
}

func AppListItems(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	wallet := r.Context().Value("wallet").(*models.Wallet)
	modelName := mux.Vars(r)["model"]

	code, settings, err := getAppSettings(app)
	if err != nil {
		api.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	var items []models.AppDataItem
	result := storage.DB.
		Order("created_at desc").
		Where(&models.AppDataItem{WalletID: wallet.ID, App: app, Model: modelName}).
		Find(&items)

	if result.Error != nil {
		api.SendJSONError(w, 500, "database error: %s", result.Error.Error())
		return
	}

	// preprocess items
	/// computed
	model := settings.getModel(modelName)
	for _, field := range model.Fields {
		if field.Computed != nil {
			for _, item := range items {
				item.Value[field.Name], _ = runlua.RunLua(runlua.Params{
					AppID:   app,
					AppCode: code,
					FunctionToRun: fmt.Sprintf(
						"get_model_field('%s', '%s').computed(item)",
						model.Name, field.Name,
					),
					InjectedGlobals: &map[string]interface{}{"item": item.Value},
				})
			}
		}
	}
	/// filter
	if model.Filter != nil {
		filteredItems := make([]models.AppDataItem, 0, len(items))
		for _, item := range items {
			returnedValue, _ := runlua.RunLua(runlua.Params{
				AppID:           app,
				AppCode:         code,
				FunctionToRun:   fmt.Sprintf("get_model('%s').filter(item)", model.Name),
				InjectedGlobals: &map[string]interface{}{"item": item.Value},
			})

			if shouldKeep, ok := returnedValue.(bool); ok && shouldKeep {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	json.NewEncoder(w).Encode(items)
}

func AppSetItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var value map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		api.SendJSONError(w, 400, "failed to read data: %s", err.Error())
		return
	}

	if err := db.Set(wallet.ID, app, key, value); err != nil {
		api.SendJSONError(w, 500, "failed to set item: %s", err.Error())
		return
	}
}

func AppDeleteItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if err := db.Delete(wallet.ID, app, key); err != nil {
		api.SendJSONError(w, 500, "failed to delete item: %s", err.Error())
		return
	}
}

func AppCustomAction(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appidToURL(mux.Vars(r)["appid"])
	action := mux.Vars(r)["action"]

	var params interface{}
	json.NewDecoder(r.Body).Decode(&params)

	code, settings, err := getAppSettings(app)
	if err != nil {
		api.SendJSONError(w, 400, "failed to get app settings: %s", err.Error())
		return
	}

	if _, ok := settings.Actions[action]; !ok {
		api.SendJSONError(w, 404, "action '%s' not defined on app: %s", action, err.Error())
		return
	}

	returned, err := runlua.RunLua(runlua.Params{
		AppID:           app,
		AppCode:         code,
		FunctionToRun:   fmt.Sprintf("actions.%s(params)", action),
		InjectedGlobals: &map[string]interface{}{"params": params},
		WalletID:        walletID,
	})
	if err != nil {
		api.SendJSONError(w, 470, "failed to run action: %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(returned)
}
