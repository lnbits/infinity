package lightning

import (
	"log"

	"github.com/lnbits/relampago"
	"github.com/lnbits/relampago/sparko"
	"github.com/lnbits/relampago/void"
	"github.com/kelseyhightower/envconfig"
	"github.com/lnbits/lnbits/events"
)

var LN relampago.Wallet

type LightningBackendSettings struct {
	SparkoURL   string `envconfig:"SPARKO_URL"`
	SparkoToken string `envconfig:"SPARKO_TOKEN"`
}

func Connect(backendType string) {
	var lbs LightningBackendSettings
	envconfig.Process("", &lbs)

	// start lightning backend
	switch backendType {
	case "lndrest":
	case "lndgrpc":
	case "eclair":
	case "clightning":
	case "sparko":
		LN = sparko.Start(sparko.Params{
			Host:               lbs.SparkoURL,
			Key:                lbs.SparkoToken,
			InvoiceLabelPrefix: "lbs",
		})
	case "lnbits":
	default:
		// use void wallet that does nothing
		LN = void.Start()
	}

	paymentsStream, err := LN.PaymentsStream()
	if err != nil {
		log.Fatalf("failed to start lightning payments stream: %s", err.Error())
	}

	paidInvoicesStream, err := LN.PaidInvoicesStream()
	if err != nil {
		log.Fatalf("failed to start lightning invoices stream: %s", err.Error())
	}

	go func() {
		for payment := range paymentsStream {
			events.NotifyPaymentSentStatus(payment)
		}
	}()

	go func() {
		for invoice := range paidInvoicesStream {
			events.NotifyInvoicePaid(invoice)
		}
	}()
}
