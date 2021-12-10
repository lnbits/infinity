package apps

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/lnbits/lnbits/events"
	"github.com/lnbits/lnbits/models"
	"github.com/rs/zerolog"
)

var AppCacheSize int
var ServerName string

var httpClient = &http.Client{
	Timeout: time.Second * 2,
	Transport: httpcache.NewTransport(
		diskcache.New("/tmp/lnbits-infinity-cache/"),
	),
	CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return fmt.Errorf("target '%s' has returned a redirect", r.URL)
	},
}

var log zerolog.Logger

var appStreams = sync.Map{}
var publicAppStreams = sync.Map{}

var nameValidator = regexp.MustCompile("^[a-z_0-9]+$")

func init() {
	go func() {
		c := make(chan models.Payment)
		events.OnPaymentReceived(c)
		for payment := range c {
			go TriggerPaymentEvent("payment_received", payment)
		}
	}()

	go func() {
		c := make(chan models.Payment)
		events.OnPaymentSent(c)
		for payment := range c {
			go TriggerPaymentEvent("payment_sent", payment)
		}
	}()

	// periodically trigger apps
	hourly := time.NewTicker(time.Hour * 1)
	go func() {
		for {
			now := <-hourly.C

			go TriggerGenericEvent("hourly", now.Unix())
			if now.Hour() == 0 {
				go TriggerGenericEvent("daily", now.Unix())
				if now.Weekday() == time.Sunday {
					go TriggerGenericEvent("weekly", now.Unix())
				}
			}
		}

		hourly.Stop()
	}()
}

func SetLogger(logger zerolog.Logger) {
	log = logger
}
