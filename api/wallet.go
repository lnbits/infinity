package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	lnurl "github.com/fiatjaf/go-lnurl"
	mux "github.com/gorilla/mux"
	rp "github.com/lnbits/relampago"

	"github.com/lnbits/infinity/api/apiutils"
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/services"
	"github.com/lnbits/infinity/storage"
	"github.com/lnbits/infinity/utils"
)

func Wallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	// load wallet balance
	wallet.Balance, _ = services.LoadWalletBalance(wallet.ID)

	// load wallet payments
	wallet.Payments, _ = services.LoadWalletPayments(wallet.ID)

	// load wallet balanceChecks
	storage.DB.Where("wallet_id = ?", wallet.ID).Find(&wallet.BalanceChecks)

	// LNURL drain URL
	wallet.LNURLDrain, _ = lnurl.LNURLEncode(
		r.URL.Scheme + "://" + r.Host + "/lnurl/wallet/drain?api-key=" + wallet.AdminKey)

	apiutils.SendJSON(w, wallet)
}

func RenameWallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if r.Context().Value("permission").(string) != "admin" {
		w.WriteHeader(401)
		return
	}

	wallet.Name = mux.Vars(r)["new-name"]
	storage.DB.Save(&wallet)

	w.WriteHeader(200)
}

func DeleteWallet(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if r.Context().Value("permission").(string) != "admin" {
		w.WriteHeader(401)
		return
	}

	wallet.AdminKey = "del:" + wallet.AdminKey
	wallet.InvoiceKey = "del:" + wallet.InvoiceKey
	wallet.UserID = "del:" + wallet.UserID

	storage.DB.Save(&wallet)

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
		apiutils.SendJSONError(w, 400, err.Error())
		return
	}

	// lnbits compatibility
	if params.Memo != "" && params.Description == "" {
		params.Description = params.Memo
	}

	if params.Description == "" {
		apiutils.SendJSONError(w, 400, "Missing description.")
		return
	}

	// transform input
	if params.DescriptionHashHex != "" && len(params.DescriptionHash) == 0 {
		params.DescriptionHash, _ = hex.DecodeString(params.DescriptionHashHex)
	}

	if params.Unit == "sat" {
		params.Msatoshi = int64(params.Amount) * 1000
	} else {
		if msats, err := utils.GetMsatsPerFiatUnit(params.Unit); err == nil {
			params.Msatoshi = int64(params.Amount * float64(msats))
		} else {
			apiutils.SendJSONError(w, 400, fmt.Sprintf("failed to get rate for currency %s: %s", params.Unit, err.Error()))
			return
		}
	}

	payment, err := services.CreateInvoice(wallet.ID, params.CreateInvoiceParams)
	if err != nil {
		apiutils.SendJSONError(w, 450, fmt.Sprintf("failed to create invoice: %s", err.Error()))
		return
	}

	apiutils.SendJSON(w, &payment)
}

func PayInvoice(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if r.Context().Value("permission").(string) != "admin" {
		w.WriteHeader(401)
		return
	}

	var params struct {
		services.PayInvoiceParams

		// lnbits compatibility
		Bolt11 string `json:"bolt11"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		apiutils.SendJSONError(w, 400, err.Error())
		return
	}

	// lnbits compatibility
	if params.Invoice == "" && params.Bolt11 != "" {
		params.Invoice = params.Bolt11
	}

	payment, err := services.PayInvoice(wallet.ID, params.PayInvoiceParams)
	if err != nil {
		apiutils.SendJSONError(w, 450, fmt.Sprintf("failed to pay invoice: %s", err.Error()))
		return
	}

	apiutils.SendJSON(w, payment)
}

func LnurlAuth(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if r.Context().Value("permission").(string) != "admin" {
		w.WriteHeader(401)
		return
	}

	var params struct {
		Callback string `json:"callback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apiutils.SendJSONError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	parsed, err := url.Parse(params.Callback)
	if err != nil {
		apiutils.SendJSONError(w, 400, "got invalid callback URL '%s': %s",
			params.Callback, err.Error())
		return
	}

	k1hex := parsed.Query().Get("k1")
	k1, err := hex.DecodeString(k1hex)
	if err != nil {
		apiutils.SendJSONError(w, 400, "Invalid k1 hex '%s': %s.", k1hex, err.Error())
		return
	}

	sk := services.AuthKey(wallet.ID, parsed.Host)
	if err := utils.PerformKeyAuthFlow(sk, parsed, k1); err != nil {
		apiutils.SendJSONError(w, 500, "Failed to sign: %s.", err.Error())
		return
	}

	w.WriteHeader(200)
}

func PayLnurl(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	if r.Context().Value("permission").(string) != "admin" {
		w.WriteHeader(401)
		return
	}

	var params struct {
		Params  lnurl.LNURLPayParams `json:"params"`
		Amount  int64                `json:"amount"`
		Comment string               `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		apiutils.SendJSONError(w, 400, "got invalid JSON: %s", err.Error())
		return
	}

	values, err := params.Params.Call(params.Amount, params.Comment, nil)
	if err != nil {
		apiutils.SendJSONError(w, 450, "failed to get lnurl invoice: %s", err.Error())
		return
	}

	extra := make(models.JSONObject)

	// store successAction
	if values.SuccessAction != nil {
		extra["success_action"] = values.SuccessAction
	}

	// store comment
	if params.Comment != "" {
		extra["comment"] = params.Comment
	}

	// actually pay
	payment, err := services.PayInvoice(wallet.ID, services.PayInvoiceParams{
		PaymentParams: rp.PaymentParams{
			Invoice: values.PR,
		},
		Extra: extra,
	})
	if err != nil {
		apiutils.SendJSONError(w, 500, "failed to pay: %s", err.Error())
		return
	}

	apiutils.SendJSON(w, struct {
		SuccessAction *lnurl.SuccessAction `json:"success_action"`
		PaymentHash   string               `json:"payment_hash"`
		CheckingID    string               `json:"checking_id"`
	}{
		values.SuccessAction,
		payment.CheckingID,
		values.ParsedInvoice.PaymentHash,
	})
}

func GetPayment(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)
	id := mux.Vars(r)["id"]

	payment := models.Payment{CheckingID: id, WalletID: wallet.ID}
	storage.DB.Where(&payment).First(&payment)

	apiutils.SendJSON(w, payment)
}

func LnurlScan(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)
	code := mux.Vars(r)["code"]

	_, lnurlParams, err := lnurl.HandleLNURL(code)
	if err != nil {
		if lnurlError, ok := err.(lnurl.LNURLErrorResponse); ok {
			w.Header().Set("Content-Type", "application/json")
			b, _ := utils.JSONMarshal(struct {
				Message string `json:"message"`
				Domain  string `json:"domain"`
			}{lnurlError.Reason, lnurlError.URL.Host})
			http.Error(w, string(b), 410)
		} else {
			apiutils.SendJSONError(w, 480, err.Error())
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
	case lnurl.LNURLPayParams:
		response.Kind = "pay"
		response.Fixed = params.MinSendable == params.MaxSendable

		h := params.HashMetadata()
		response.DescriptionHashHex = hex.EncodeToString(h[:])

		response.Description = params.Metadata.Description
		response.Image = params.Metadata.Image.DataURI
		response.TargetUser = params.Metadata.LightningAddress
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
			services.AuthKey(wallet.ID, params.CallbackURL.Host).
				PubKey().SerializeCompressed(),
		)
	default:
		apiutils.SendJSONError(w, 400, "Unsupported LNURL.")
		return
	}

	apiutils.SendJSON(w, response)
}
