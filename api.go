package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	lnurl "github.com/fiatjaf/go-lnurl"
	decodepay "github.com/fiatjaf/ln-decodepay"
	rp "github.com/fiatjaf/relampago"
	mux "github.com/gorilla/mux"
)

func apiUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)

	// load wallets
	db.Raw(`
      SELECT *,
        (SELECT sum(amount) FROM payments AS p
         WHERE p.wallet_id = w.id
           AND ( amount < 0
            OR   ( amount > 0 AND NOT pending )
               )
        ) AS balance FROM wallets AS w
      WHERE w.user_id = ?
    `, user.ID).Scan(&user.Wallets)

	json.NewEncoder(w).Encode(user)
}

func apiCreateWallet(w http.ResponseWriter, r *http.Request) {
	var masterKey string
	user := &User{}

	if r.Context().Value("user") != nil {
		user = r.Context().Value("user").(*User)
	} else {
		// create user
		user.ID = randomHex(16)
		user.Apps = make(StringList, 0)
		masterKey = randomHex(32) // will only be returned if we're creating the user
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

	wallet := Wallet{
		ID:         randomHex(16),
		Name:       params.Name,
		UserID:     user.ID,
		InvoiceKey: randomHex(32),
		AdminKey:   randomHex(32),
	}
	result := db.Create(&wallet)
	if result.Error != nil {
		jsonError(w, 400, "error saving wallet: %s", result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(struct {
		UserMasterKey string `json:"userMasterKey"`
		Wallet        Wallet `json:"wallet"`
	}{
		masterKey,
		wallet,
	})
}

func apiWallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)

	// load wallet balance
	db.Model(&Payment{}).
		Select("sum(amount)").
		Where("amount < 0 OR (amount > 0 AND NOT pending)").
		Where("wallet_id = ?", wallet.ID).
		First(&wallet.Balance)

	// load wallet payments
	db.Where("wallet_id = ?", wallet.ID).Find(&wallet.Payments)

	// load wallet balanceChecks
	db.Where("wallet_id = ?", wallet.ID).Find(&wallet.BalanceChecks)

	json.NewEncoder(w).Encode(wallet)
}

func apiRenameWallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)

	wallet.Name = mux.Vars(r)["new-name"]
	db.Save(&wallet)

	w.WriteHeader(200)
}

func apiCreateInvoice(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)

	var params struct {
		*CreateInvoiceParams

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
		if msats, err := getMsatsPerFiatUnit(params.Unit); err == nil {
			params.Msatoshi = int64(params.Amount * float64(msats))
		} else {
			jsonError(w, 400, fmt.Sprintf("failed to get rate for currency %s: %s", params.Unit, err.Error()))
			return
		}
	}

	payment, err := wallet.CreateInvoice(params.CreateInvoiceParams)
	if err != nil {
		jsonError(w, 450, fmt.Sprintf("failed to create invoice: %s", err.Error()))
		return
	}

	json.NewEncoder(w).Encode(&payment)
}

func apiPayInvoice(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)

	var params struct {
		*PayInvoiceParams

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

	payment, err := wallet.PayInvoice(params.PayInvoiceParams)
	if err != nil {
		jsonError(w, 450, fmt.Sprintf("failed to pay invoice: %s", err.Error()))
		return
	}

	json.NewEncoder(w).Encode(payment)
}

func apiLnurlAuth(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)

	log.Print(wallet)
}

func apiPayLnurl(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)

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
		jsonError(w, 420, "%s has sent an invalid invoice")
		return
	}
	if inv.DescriptionHash != params.DescriptionHashHex {
		jsonError(w, 420, "%s has sent an invoice with wrong description_hash")
		return
	}
	if int64(inv.MSatoshi) != params.Amount {
		jsonError(w, 420, "%s has sent an invoice with wrong msatoshi amount")
		return
	}

	extra := make(JSONObject)

	// store successAction
	if lnurlResponse.SuccessAction != nil {
		extra["success_action"] = lnurlResponse.SuccessAction
	}

	// store comment
	if params.Comment != "" {
		extra["comment"] = params.Comment
	}

	// actually pay
	payment, err := wallet.PayInvoice(&PayInvoiceParams{
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

func apiGetPayment(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*Wallet)
	id := mux.Vars(r)["id"]

	payment := Payment{CheckingID: id, WalletID: wallet.ID}
	db.Where(&payment).First(&payment)

	json.NewEncoder(w).Encode(payment)
}
