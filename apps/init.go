package apps

import (
	"net/http"
	"sync"
	"time"

	"github.com/lnbits/lnbits/events"
	"github.com/lnbits/lnbits/models"
	"github.com/rs/zerolog"
)

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

var log zerolog.Logger

var appStreams = sync.Map{}

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

func SetLogger(logger zerolog.Logger) {
	log = logger
}
