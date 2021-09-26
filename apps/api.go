package apps

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"

	api "github.com/lnbits/lnbits/api"
	models "github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func AppInfo(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])

	code, err := getAppCode(app)
	if err != nil {
		api.SendJSONError(w, 420, "failed to fetch app code from %s: %s", app, err.Error())
		return
	}

	settings, err := getSettings(code)
	if err != nil {
		api.SendJSONError(w, 400, "app init() didn't return settings: %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(settings)
}

func AppListItems(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var items []models.AppDataItem
	if result := storage.DB.
		Where(&models.AppDataItem{WalletID: wallet.ID, App: app}).
		Find(&items); result.Error != nil {
		api.SendJSONError(w, 500, "database error: %s", result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(items)
}

func AppSetItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	item := models.AppDataItem{
		WalletID: wallet.ID,
		App:      app,
		Key:      key,
	}
	if err := json.NewDecoder(r.Body).Decode(&item.Value); err != nil {
		api.SendJSONError(w, 400, "failed to read data: %s", err.Error())
		return
	}

	storage.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app"}, {Name: "wallet_id"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&item)
}

func AppDeleteItem(w http.ResponseWriter, r *http.Request) {
	app := appidToURL(mux.Vars(r)["appid"])
	key := mux.Vars(r)["key"]
	wallet := r.Context().Value("wallet").(*models.Wallet)

	storage.DB.Delete(&models.AppDataItem{
		WalletID: wallet.ID,
		App:      app,
		Key:      key,
	})
}

func AppCustom(w http.ResponseWriter, r *http.Request) {
	walletID := mux.Vars(r)["wallet"]
	app := appidToURL(mux.Vars(r)["appid"])

	log.Print(walletID, app)
}
