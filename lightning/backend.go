package lightning

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/lnbits/lnbits/events"
	"github.com/lnbits/relampago"
	"github.com/lnbits/relampago/lnd"
	"github.com/lnbits/relampago/sparko"
	"github.com/lnbits/relampago/void"
)

var LN relampago.Wallet

type LightningBackendSettings struct {
	SparkoURL   string `envconfig:"SPARKO_URL"`
	SparkoToken string `envconfig:"SPARKO_TOKEN"`

	LNDHost         string `envconfig:"LND_HOST"`
	LNDCertPath     string `envconfig:"LND_CERT_PATH"`
	LNDMacaroonPath string `envconfig:"LND_MACAROON_PATH"`
}

func Connect(backendType string) {
	var lbs LightningBackendSettings
	envconfig.Process("", &lbs)

	// start lightning backend
	var err error
	switch backendType {
	case "lndrest":
	case "lndgrpc":
		LN, err = lnd.Start(lnd.Params{
			Host:         lbs.LNDHost,
			CertPath:     lbs.LNDCertPath,
			MacaroonPath: lbs.LNDMacaroonPath,
		})
	case "eclair":
	case "clightning":
	case "sparko":
		LN, err = sparko.Start(sparko.Params{
			Host:               lbs.SparkoURL,
			Key:                lbs.SparkoToken,
			InvoiceLabelPrefix: "lbs",
		})
	case "lnbits":
	default:
		// use void wallet that does nothing
		LN, err = void.Start()
	}
	if err != nil {
		log.Fatalf("failed to initialize %s backend with %v: %s", backendType, lbs, err)
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
