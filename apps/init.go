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

	go func() {
		c := make(chan events.GenericEvent)
		events.OnGenericEvent(c)
		for genericEvent := range c {
			if genericEvent.App != "" && genericEvent.Wallet != "" {
				go TriggerEventOnSpecificAppWallet(
					AppWallet{genericEvent.Wallet, genericEvent.App},
					genericEvent.Name,
					genericEvent.Data,
				)
			} else {
				go TriggerGlobalEvent(genericEvent.Name, genericEvent.Data)
			}
		}
	}()

	// trigger an event when lnbits starts
	go func() {
		time.Sleep(3 * time.Second)
		TriggerGlobalEvent("init", nil)
	}()

	// periodically trigger apps
	hourly := time.NewTicker(time.Hour * 1)
	go func() {
		for {
			now := <-hourly.C

			go TriggerGlobalEvent("hourly", now.Unix())
			if now.Hour() == 0 {
				go TriggerGlobalEvent("daily", now.Unix())
				if now.Weekday() == time.Sunday {
					go TriggerGlobalEvent("weekly", now.Unix())
				}
			}
		}

		hourly.Stop()
	}()
}

func SetLogger(logger zerolog.Logger) {
	log = logger
}
