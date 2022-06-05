package api

import (
	"encoding/json"
	"net/http"

	"github.com/lnbits/infinity/api/apiutils"
	"github.com/lnbits/infinity/apps"
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/services"
	"github.com/lnbits/infinity/storage"
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

	// load apps
	storage.DB.Raw("SELECT url FROM user_apps WHERE user_id = ?", user.ID).
		Scan(&user.Apps)

	apiutils.SendJSON(w, user)
}

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apiutils.SendJSONError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	var masterKey string
	user := &models.User{}

	if r.Context().Value("user") != nil {
		user = r.Context().Value("user").(*models.User)
	} else {
		// create user
		var err error
		user, err = services.CreateUser()
		if err != nil {
			apiutils.SendJSONError(w, 500, "error saving user: %s", err.Error())
			return
		}
		masterKey = user.MasterKey
	}

	// create wallet
	wallet, err := services.CreateWallet(user.ID, params.Name)
	if err != nil {
		apiutils.SendJSONError(w, 500, "error saving wallet: %s", err.Error())
		return
	}

	apiutils.SendJSON(w, struct {
		UserMasterKey string         `json:"userMasterKey"`
		Wallet        *models.Wallet `json:"wallet"`
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
	if _, err := apps.GetAppSettings(params.URL, true); err != nil {
		apiutils.SendJSONError(w, 470, "failed to run app: %s", err.Error())
		return
	}

	// add it to the list of apps for this user
	if resp := storage.DB.Create(&models.UserApp{
		UserID: user.ID,
		URL:    params.URL,
	}); resp.Error != nil {
		apiutils.SendJSONError(w, 500, "failed to save app: %s", resp.Error.Error())
		return
	}

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

	result := storage.DB.Where("url = ? AND user_id = ?", params.URL, user.ID).
		Delete(&models.UserApp{})

	if result.Error != nil {
		apiutils.SendJSONError(w, 500, "failed to delete app: %s", result.Error.Error())
		return
	}

	w.WriteHeader(200)
}
