package api

import (
	"fmt"
	"net/http"

	lnurl "github.com/fiatjaf/go-lnurl"
	rp "github.com/lnbits/relampago"
	"github.com/lnbits/lnbits/api/apiutils"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/services"
	"github.com/lnbits/lnbits/storage"
)

func DrainFunds(w http.ResponseWriter, r *http.Request) {
	walletKey := r.URL.Query().Get("api-key")

	// only allow admin keys
	var wallet *models.Wallet
	result := storage.DB.Where("admin_key", walletKey).First(&wallet)
	if result.Error != nil {
		apiutils.SendJSON(w, lnurl.LNURLErrorResponse{
			Status: "ERROR",
			Reason: "Can't withdraw. Invalid API key.",
		})
		return
	}

	if pr := r.URL.Query().Get("pr"); pr != "" {
		// this is the callback already
		// save balanceNotify
		if bn := r.URL.Query().Get("balanceNotify"); bn != "" {
			storage.DB.Model(&models.Wallet{}).
				Where("id", wallet.ID).
				Update("balance_notify", bn)
		}

		// pay invoice
		_, err := services.PayInvoice(wallet.ID, services.PayInvoiceParams{
			PaymentParams: rp.PaymentParams{Invoice: pr},
			Tag:           "drain",
		})
		if err != nil {
			apiutils.SendJSON(w, lnurl.LNURLResponse{
				Status: "OK",
			})
			return
		}
	} else {
		// return lnurl-withdraw params
		// load wallet balance
		balance, _ := services.LoadWalletBalance(wallet.ID)
		thisURL := r.Host + "/lnurl/wallet/drain?api-key=" + walletKey
		response := lnurl.LNURLWithdrawResponse{
			Tag:      "withdrawRequest",
			Callback: thisURL,
			K1:       "0",
			DefaultDescription: fmt.Sprintf("balance withdraw from %s @ %s",
				wallet.Name, SiteTitle),
			BalanceCheck: thisURL,
		}

		if balance > 1000 {
			response.MinWithdrawable = 1000
			response.MaxWithdrawable = balance
		} else {
			response.MinWithdrawable = 0
			response.MaxWithdrawable = 0
		}

		apiutils.SendJSON(w, response)
	}
}
