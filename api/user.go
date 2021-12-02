package api

import (
	"encoding/json"
	"net/http"

	"github.com/lnbits/lnbits/api/apiutils"
	"github.com/lnbits/lnbits/apps"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/lnbits/lnbits/utils"
	"github.com/lucsky/cuid"
)

func User(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	// load wallets
	storage.DB.Raw(`
      SELECT *,
        (SELECT coalesce(sum(amount), 0) FROM payments AS p
         WHERE p.wallet_id = w.id
           AND ( amount < 0
            OR   ( amount > 0 AND NOT pending )
               )
        ) AS balance FROM wallets AS w
      WHERE w.user_id = ?
    `, user.ID).Scan(&user.Wallets)

	apiutils.SendJSON(w, user)
}

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	var masterKey string
	user := &models.User{}

	if r.Context().Value("user") != nil {
		user = r.Context().Value("user").(*models.User)
	} else {
		// create user
		user.ID = cuid.Slug()
		user.Apps = make(models.StringList, 0)
		masterKey = utils.RandomHex(32) // will only be returned if we're creating the user
		user.MasterKey = masterKey
		storage.DB.Create(user)
	}

	// create wallet
	var params struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apiutils.SendJSONError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	wallet := models.Wallet{
		ID:         cuid.Slug(),
		Name:       params.Name,
		UserID:     user.ID,
		InvoiceKey: utils.RandomHex(32),
		AdminKey:   utils.RandomHex(32),
	}
	result := storage.DB.Create(&wallet)
	if result.Error != nil {
		apiutils.SendJSONError(w, 400, "error saving wallet: %s", result.Error.Error())
		return
	}

	apiutils.SendJSON(w, struct {
		UserMasterKey string        `json:"userMasterKey"`
		Wallet        models.Wallet `json:"wallet"`
	}{
		masterKey,
		wallet,
	})
}

func AddApp(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var params struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apiutils.SendJSONError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	// try to fetch settings for this app first
	if _, err := apps.GetAppSettings(params.URL); err != nil {
		apiutils.SendJSONError(w, 470, "failed to run app: %s", err.Error())
		return
	}

	// add it to the list of apps for this user
	for _, app := range user.Apps {
		if app == params.URL {
			w.WriteHeader(200)
			return
		}
	}
	user.Apps = append(user.Apps, params.URL)
	storage.DB.Save(user)

	w.WriteHeader(201)
}

func RemoveApp(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	var params struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apiutils.SendJSONError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	newAppsList := make([]string, 0, len(user.Apps))
	for _, app := range user.Apps {
		if app != params.URL {
			newAppsList = append(newAppsList, app)
		}
	}
	user.Apps = newAppsList
	storage.DB.Save(user)

	w.WriteHeader(200)
}
