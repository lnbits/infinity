package apps

import (
	"net/http"
	"time"

	"github.com/lnbits/lnbits/events"
	"github.com/lnbits/lnbits/models"
)

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

func init() {
	go func() {
		c := make(chan models.Payment)
		events.OnPaymentReceived(c)
		for payment := range c {
			go TriggerEvent("payment_received", payment)
		}
	}()

	go func() {
		c := make(chan models.Payment)
		events.OnPaymentSent(c)
		for payment := range c {
			go TriggerEvent("payment_sent", payment)
		}
	}()
}
