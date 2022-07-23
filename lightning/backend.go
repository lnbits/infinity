package lightning

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/lnbits/infinity/events"
	"github.com/lnbits/relampago"
	"github.com/lnbits/relampago/cliche"
	"github.com/lnbits/relampago/eclair"
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

	EclairHost     string `envconfig:"ECLAIR_HOST"`
	EclairPassword string `envconfig:"ECLAIR_PASSWORD"`

	ClicheJARPath    string `envconfig:"CLICHE_JAR_PATH"`
	ClicheBinaryPath string `envconfig:"CLICHE_BINARY_PATH"`
	ClicheDataDir    string `envconfig:"CLICHE_DATADIR"`
}

func Connect(backendType string) {
	var lbs LightningBackendSettings
	envconfig.Process("", &lbs)

	// start lightning backend
	var err error
	switch backendType {
	case "lndrest":
	case "lnd", "lndgrpc":
		LN, err = lnd.Start(lnd.Params{
			Host:           lbs.LNDHost,
			CertPath:       lbs.LNDCertPath,
			MacaroonPath:   lbs.LNDMacaroonPath,
			ConnectTimeout: 5 * time.Second,
		})
	case "eclair":
		LN, err = eclair.Start(eclair.Params{
			Host:     lbs.EclairHost,
			Password: lbs.EclairPassword,
		})
	case "clightning":
	case "sparko":
		LN, err = sparko.Start(sparko.Params{
			Host:               lbs.SparkoURL,
			Key:                lbs.SparkoToken,
			InvoiceLabelPrefix: "lbs",
		})
	case "cliche", "cliché":
		LN, err = cliche.Start(cliche.Params{
			JARPath:    lbs.ClicheJARPath,
			BinaryPath: lbs.ClicheBinaryPath,
			DataDir:    lbs.ClicheDataDir,
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
