package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	lnurl "github.com/fiatjaf/go-lnurl"
	decodepay "github.com/fiatjaf/ln-decodepay"
	rp "github.com/fiatjaf/relampago"
	mux "github.com/gorilla/mux"

	models "github.com/lnbits/lnbits/models"
	services "github.com/lnbits/lnbits/services"
	utils "github.com/lnbits/lnbits/utils"
)

func User(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	// load wallets
	db.Raw(`
      SELECT *,
        (SELECT coalesce(sum(amount), 0) FROM payments AS p
         WHERE p.wallet_id = w.id
           AND ( amount < 0
            OR   ( amount > 0 AND NOT pending )
               )
        ) AS balance FROM wallets AS w
      WHERE w.user_id = ?
    `, user.ID).Scan(&user.Wallets)

	json.NewEncoder(w).Encode(user)
}

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	var masterKey string
	user := &models.User{}

	if r.Context().Value("user") != nil {
		user = r.Context().Value("user").(*models.User)
	} else {
		// create user
		user.ID = utils.RandomHex(16)
		user.Apps = make(models.StringList, 0)
		masterKey = utils.RandomHex(32) // will only be returned if we're creating the user
		user.MasterKey = masterKey
		db.Create(user)
	}

	// create wallet
	var params struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		jsonError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	wallet := models.Wallet{
		ID:         utils.RandomHex(16),
		Name:       params.Name,
		UserID:     user.ID,
		InvoiceKey: utils.RandomHex(32),
		AdminKey:   utils.RandomHex(32),
	}
	result := db.Create(&wallet)
	if result.Error != nil {
		jsonError(w, 400, "error saving wallet: %s", result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(struct {
		UserMasterKey string        `json:"userMasterKey"`
		Wallet        models.Wallet `json:"wallet"`
	}{
		masterKey,
		wallet,
	})
}

func Wallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	// load wallet balance
	db.Model(&models.Payment{}).
		Select("coalesce(sum(amount), 0)").
		Where("amount < 0 OR (amount > 0 AND NOT pending)").
		Where("wallet_id = ?", wallet.ID).
		First(&wallet.Balance)

	// load wallet payments
	db.Where("wallet_id = ?", wallet.ID).Find(&wallet.Payments)

	// load wallet balanceChecks
	db.Where("wallet_id = ?", wallet.ID).Find(&wallet.BalanceChecks)

	json.NewEncoder(w).Encode(wallet)
}

func RenameWallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	wallet.Name = mux.Vars(r)["new-name"]
	db.Save(&wallet)

	w.WriteHeader(200)
}

func CreateInvoice(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var params struct {
		services.CreateInvoiceParams

		Unit               string  `json:"unit"`
		Amount             float64 `json:"amount"`
		LnurlCallback      string  `json:"lnurlCallback"`
		LnurlBalanceCheck  string  `json:"lnurlBalanceCheck"`
		DescriptionHashHex string  `json:"description_hash"`

		// lnbits compatibility
		Memo string `json:"memo"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}

	// lnbits compatibility
	if params.Memo != "" && params.Description == "" {
		params.Description = params.Memo
	}

	// transform input
	if params.DescriptionHashHex != "" && len(params.DescriptionHash) == 0 {
		params.DescriptionHash, _ = hex.DecodeString(params.DescriptionHashHex)
	}

	if params.Unit == "sat" {
		params.Msatoshi = int64(params.Amount) * 1000
	} else {
		if msats, err := services.GetMsatsPerFiatUnit(params.Unit); err == nil {
			params.Msatoshi = int64(params.Amount * float64(msats))
		} else {
			jsonError(w, 400, fmt.Sprintf("failed to get rate for currency %s: %s", params.Unit, err.Error()))
			return
		}
	}

	payment, err := services.CreateInvoice(wallet, params.CreateInvoiceParams)
	if err != nil {
		jsonError(w, 450, fmt.Sprintf("failed to create invoice: %s", err.Error()))
		return
	}

	json.NewEncoder(w).Encode(&payment)
}

func PayInvoice(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var params struct {
		services.PayInvoiceParams

		// lnbits compatibility
		Bolt11 string `json:"bolt11"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		jsonError(w, 400, err.Error())
		return
	}

	// lnbits compatibility
	if params.Invoice == "" && params.Bolt11 != "" {
		params.Invoice = params.Bolt11
	}

	payment, err := services.PayInvoice(wallet, params.PayInvoiceParams)
	if err != nil {
		jsonError(w, 450, fmt.Sprintf("failed to pay invoice: %s", err.Error()))
		return
	}

	json.NewEncoder(w).Encode(payment)
}

func LnurlAuth(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var params struct {
		Callback string `json:"callback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		jsonError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	parsed, err := url.Parse(params.Callback)
	if err != nil {
		jsonError(w, 400, "got invalid callback URL '%s': %s",
			params.Callback, err.Error())
		return
	}

	k1hex := parsed.Query().Get("k1")
	k1, err := hex.DecodeString(k1hex)
	if err != nil {
		jsonError(w, 400, "Invalid k1 hex '%s': %s.", k1hex, err.Error())
		return
	}

	sk := services.AuthKey(wallet, parsed.Host)
	if err := services.PerformKeyAuthFlow(sk, parsed, k1); err != nil {
		jsonError(w, 500, "Failed to sign: %s.", err.Error())
		return
	}

	w.WriteHeader(200)
}

func PayLnurl(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var params struct {
		DescriptionHashHex string `json:"description_hash"`
		Callback           string `json:"callback"`
		Amount             int64  `json:"amount"`
		Comment            string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		jsonError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	// call callback with params and get invoice
	callback, err := url.Parse(params.Callback)
	if err != nil {
		jsonError(w, 400, "got invalid callback URL: %s", err.Error())
		return
	}
	qs := callback.Query()
	if params.Comment != "" {
		qs.Set("comment", params.Comment)
	}
	qs.Set("amount", fmt.Sprintf("%d", params.Amount))
	callback.RawQuery = qs.Encode()

	var lnurlResponse lnurl.LNURLPayResponse2
	if resp, err := http.Get(callback.String()); err != nil {
		jsonError(w, 400, "failed to call %s: %s", callback.String(), err.Error())
		return
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&lnurlResponse); err != nil {
			jsonError(w, 400, "got invalid json from %s", callback.Host)
			return
		}
	}
	if lnurlResponse.Status == "ERROR" {
		jsonError(w, 420, "%s said: %s", callback.Host, lnurlResponse.Reason)
		return
	}

	// check invoice amount and description_hash
	inv, err := decodepay.Decodepay(lnurlResponse.PR)
	if err != nil {
		jsonError(w, 420, "%s has sent an invalid invoice", callback.Host)
		return
	}
	if inv.DescriptionHash != params.DescriptionHashHex {
		jsonError(w, 420, "%s has sent an invoice with wrong description_hash", callback.Host)
		return
	}
	if int64(inv.MSatoshi) != params.Amount {
		jsonError(w, 420, "%s has sent an invoice with wrong msatoshi amount", callback.Host)
		return
	}

	extra := make(models.JSONObject)

	// store successAction
	if lnurlResponse.SuccessAction != nil {
		extra["success_action"] = lnurlResponse.SuccessAction
	}

	// store comment
	if params.Comment != "" {
		extra["comment"] = params.Comment
	}

	// actually pay
	payment, err := services.PayInvoice(wallet, services.PayInvoiceParams{
		PaymentParams: rp.PaymentParams{
			Invoice: lnurlResponse.PR,
		},
		Extra: extra,
	})
	if err != nil {
		jsonError(w, 500, "failed to pay: %s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(struct {
		SuccessAction *lnurl.SuccessAction `json:"success_action"`
		PaymentHash   string               `json:"payment_hash"`
		CheckingID    string               `json:"checking_id"`
	}{
		lnurlResponse.SuccessAction,
		payment.CheckingID,
		inv.PaymentHash,
	})
}

func GetPayment(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)
	id := mux.Vars(r)["id"]

	payment := models.Payment{CheckingID: id, WalletID: wallet.ID}
	db.Where(&payment).First(&payment)

	json.NewEncoder(w).Encode(payment)
}

func LnurlScan(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)
	code := mux.Vars(r)["code"]

	_, lnurlParams, err := lnurl.HandleLNURL(code)
	if err != nil {
		if lnurlError, ok := err.(lnurl.LNURLErrorResponse); ok {
			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(struct {
				Message string `json:"message"`
				Domain  string `json:"domain"`
			}{lnurlError.Reason, lnurlError.URL.Host})
			http.Error(w, string(b), 410)
		} else {
			jsonError(w, 480, err.Error())
		}
	}

	var response struct {
		lnurl.LNURLParams

		Kind string `json:"kind"`

		// pay + withdraw
		Fixed bool `json:"fixed,omitempty"`

		// auth + withdraw
		Callback string `json:"callback,omitempty"`

		// pay
		DescriptionHashHex string `json:"description_hash,omitempty"`
		Description        string `json:"description,omitempty"`
		Image              string `json:"image,omitempty"`
		TargetUser         string `json:"targetUser,omitempty"`
		CommentAllowed     int    `json:"commentAllowed,omitempty"`

		// withdraw
		BalanceCheck string `json:"balanceCheck,omitempty"`

		// auth
		Pubkey string `json:"pubkey,omitempty"`
	}

	response.LNURLParams = lnurlParams

	switch params := lnurlParams.(type) {
	case lnurl.LNURLPayResponse1:
		response.Kind = "pay"
		response.Fixed = params.MinSendable == params.MaxSendable

		h := sha256.Sum256([]byte(params.EncodedMetadata))
		response.DescriptionHashHex = hex.EncodeToString(h[:])

		response.Description = params.Metadata.Description()
		response.Image = params.Metadata.ImageDataURI()
		response.TargetUser = params.Metadata.LightningAddress()
		response.CommentAllowed = int(params.CommentAllowed)
	case lnurl.LNURLWithdrawResponse:
		response.Kind = "withdraw"
		response.Fixed = params.MinWithdrawable == params.MaxWithdrawable
		response.BalanceCheck = params.BalanceCheck

		callback := params.CallbackURL
		qs := callback.Query()
		qs.Set("k1", params.K1)
		callback.RawQuery = qs.Encode()
		response.Callback = callback.String()
	case lnurl.LNURLAuthParams:
		response.Kind = "auth"
		response.Pubkey = hex.EncodeToString(
			services.AuthKey(wallet, params.CallbackURL.Host).
				PubKey().SerializeCompressed(),
		)
	default:
		jsonError(w, 400, "Unsupported LNURL.")
		return
	}

	json.NewEncoder(w).Encode(response)
}
