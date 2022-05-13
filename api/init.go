package api

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lnbits/infinity/events"
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
	"github.com/lnbits/infinity/utils"
)

var walletStreams = sync.Map{}

var SiteTitle string

var webhookClient = &http.Client{
	Timeout: time.Second * 2,
	CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return fmt.Errorf("target '%s' has returned a redirect", r.URL)
	},
}

func init() {
	go func() {
		c := make(chan models.Payment)
		events.OnPaymentReceived(c)
		for payment := range c {
			// sse stream
			SendWalletSSE(payment.WalletID, "payment-received", payment)

			// webhook
			if payment.Webhook != "" && payment.WebhookStatus == 0 {
				j, _ := utils.JSONMarshal(payment)
				b := bytes.NewBuffer(j)
				resp, err := webhookClient.Post(payment.Webhook, "application/json", b)
				var status int
				if err != nil {
					status = -1
				} else {
					status = resp.StatusCode
				}
				storage.DB.
					Model(&payment).
					Where(&payment).
					Update("webhook_status", status)
			}

			// balanceNotify
			var wallet models.Wallet
			storage.DB.
				Where("id = ?", payment.WalletID).
				Where("balance_notify IS NOT NULL").
				First(&wallet)
			if wallet.BalanceNotify != "" {
				go webhookClient.Post(wallet.BalanceNotify, "application/lnurl", nil)
			}
		}
	}()

	go func() {
		c := make(chan models.Payment)
		events.OnPaymentSent(c)
		for payment := range c {
			// sse stream
			go SendWalletSSE(payment.WalletID, "payment-sent", payment)

			// webhook
			if payment.Webhook != "" && payment.WebhookStatus == 0 {
				j, _ := utils.JSONMarshal(payment)
				b := bytes.NewBuffer(j)
				resp, err := webhookClient.Post(payment.Webhook, "application/json", b)
				var status int
				if err != nil {
					status = -1
				} else {
					status = resp.StatusCode
				}
				storage.DB.
					Model(&payment).
					Where(&payment).
					Update("webhook_status", status)
			}
		}
	}()

	go func() {
		c := make(chan models.Payment)
		events.OnPaymentSent(c)
		for payment := range c {
			go SendWalletSSE(payment.WalletID, "payment-failed", payment.CheckingID)
		}
	}()
}
