package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lnbits/lnbits/models"
	"gopkg.in/antage/eventsource.v1"
)

func SSE(w http.ResponseWriter, r *http.Request) {
	wallet := r.Context().Value("wallet").(*models.Wallet)

	var es eventsource.EventSource
	ies, ok := walletStreams.Load(wallet.ID)

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

		walletStreams.Store(wallet.ID, es)
	} else {
		es = ies.(eventsource.EventSource)
	}

	go func() {
		time.Sleep(1 * time.Second)
		es.SendRetryMessage(3 * time.Second)
	}()

	es.ServeHTTP(w, r)
}

func SendWalletSSE(walletID string, typ string, data interface{}) {
	jpayload, _ := json.Marshal(data)
	payload := string(jpayload)

	if ies, ok := walletStreams.Load(walletID); ok {
		ies.(eventsource.EventSource).SendEventMessage(payload, typ, "")
	}
}
