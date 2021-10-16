package api

import (
	"sync"

	"github.com/lnbits/lnbits/events"
	"github.com/lnbits/lnbits/models"
)

var walletStreams = sync.Map{}

func init() {
	go func() {
		c := make(chan models.Payment)
		events.OnPaymentReceived(c)
		for payment := range c {
			SendWalletSSE(payment.WalletID, "payment-received", payment)
		}
	}()

	go func() {
		c := make(chan models.Payment)
		events.OnPaymentSent(c)
		for payment := range c {
			go SendWalletSSE(payment.WalletID, "payment-sent", payment)
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
