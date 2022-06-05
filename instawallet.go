package main

import (
	"fmt"
	"net/http"

	"github.com/fiatjaf/go-lnurl"
	"github.com/lnbits/infinity/services"
	rp "github.com/lnbits/relampago"
)

func instawallet(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("lightning")

	raw, lnurlParams, err := lnurl.HandleLNURL(code)
	if err != nil {
		if raw == "" {
			raw = "~"
		}
		http.Error(w, fmt.Sprintf("failed to fetch lnurl voucher params from %s (%s): %s", code, raw, err.Error()), 400)
		return
	}

	if params, ok := lnurlParams.(lnurl.LNURLWithdrawResponse); !ok {
		http.Error(w, fmt.Sprintf("lnurl '%s' is not a valid lnurl-withdraw voucher that can be claimed", raw), 400)
		return
	} else {
		// create user and wallet
		user, err := services.CreateUser()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create user: %s", err.Error()), 500)
			return
		}

		wallet, err := services.CreateWallet(user.ID, "from-voucher")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create wallet: %s", err.Error()), 500)
			return
		}

		// create invoice
		invoice, err := services.CreateInvoice(wallet.ID, services.CreateInvoiceParams{
			InvoiceParams: rp.InvoiceParams{
				Msatoshi:    params.MaxWithdrawable,
				Description: params.DefaultDescription,
			},
			Tag: "voucher-claim",
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create invoice: %s", err.Error()), 500)
			return
		}

		// send invoice to lnurl-withdraw callback
		callback := params.CallbackURL
		qs := callback.Query()
		qs.Set("k1", params.K1)
		qs.Set("pr", invoice.Bolt11)
		callback.RawQuery = qs.Encode()
		http.Get(callback.String())

		// redirect to wallet interface
		http.Redirect(w, r,
			fmt.Sprintf("%s/wallet/%s?key=%s", s.ServiceURL, wallet.ID, user.MasterKey),
			http.StatusFound)
	}
}
