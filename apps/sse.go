package apps

import (
	"net/http"
	"time"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/utils"
	"gopkg.in/antage/eventsource.v1"
)

func SSE(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var es eventsource.EventSource
	ies, ok := appStreams.Load(wallet.ID)

	if !ok {
		es = eventsource.New(
			&eventsource.Settings{
				Timeout:        5 * time.Second,
				CloseOnTimeout: true,
				IdleTimeout:    1 * time.Minute,
			},
			func(r *http.Request) [][]byte {
				return [][]byte{
					[]byte("X-Accel-Buffering: no"),
					[]byte("Cache-Control: no-cache"),
					[]byte("Content-Type: text/event-stream"),
					[]byte("Connection: keep-alive"),
					[]byte("Access-Control-Allow-Origin: *"),
				}
			},
		)
		go func() {
			for {
				time.Sleep(25 * time.Second)
				es.SendEventMessage("", "keepalive", "")
			}
		}()

		appStreams.Store(wallet.ID, es)
	} else {
		es = ies.(eventsource.EventSource)
	}

	go func() {
		time.Sleep(1 * time.Second)
		es.SendRetryMessage(3 * time.Second)
	}()

	es.ServeHTTP(w, r)
}

func SendItemSSE(item models.AppDataItem) {
	jpayload, _ := utils.JSONMarshal(item)
	payload := string(jpayload)

	if ies, ok := appStreams.Load(item.WalletID); ok {
		ies.(eventsource.EventSource).SendEventMessage(payload, "item", "")
	}
}

func SendPrintSSE(walletID string, prints []byte) {
	if ies, ok := appStreams.Load(walletID); ok {
		ies.(eventsource.EventSource).SendEventMessage(string(prints[:]), "print", "")
	}
}
