package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fiatjaf/relampago"
	"github.com/fiatjaf/relampago/sparko"
	"github.com/fiatjaf/relampago/void"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Settings struct {
	Host     string `envconfig:"HOST" default:"0.0.0.0"`
	Port     string `envconfig:"PORT" default:"5000"`
	Database string `envconfig:"DATABASE" required:"true"`

	SiteTitle         string   `envconfig:"LNBITS_SITE_TITLE" default:"LNBitsLocal"`
	SiteTagline       string   `envconfig:"LNBITS_SITE_TAGLINE" default:"Locally-hosted lightning wallet"`
	SiteDescription   string   `envconfig:"LNBITS_SITE_DESCRIPTION" default:""`
	ThemeOptions      []string `envconfig:"LNBITS_THEME_OPTIONS" default:"classic, flamingo, mint, salvador, monochrome, autumn"`
	DefaultWalletName string   `envconfig:"LNBITS_DEFAULT_WALLET_NAME" default:"LNbits Wallet"`

	LightningBackend string `envconfig:"LNBITS_LIGHTNING_BACKEND" default:"void"`
	SparkoURL        string `envconfig:"SPARKO_URL"`
	SparkoToken      string `envconfig:"SPARKO_TOKEN"`
}

var s Settings
var ln relampago.Wallet
var db *gorm.DB
var log = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stdout})
var router = mux.NewRouter()
var commit string // will be set at compile time

//go:embed client/dist/spa
var static embed.FS

func main() {
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig.")
		return
	}

	// setup logger
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log = log.With().Timestamp().Logger()

	// database
	opts := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}
	if strings.HasPrefix(s.Database, "postgres") {
		// postgres
		db, err = gorm.Open(postgres.Open(s.Database), opts)
	} else if strings.HasPrefix(s.Database, "cockroach") {
		// cockroach
		db, err = gorm.Open(postgres.Open(s.Database), opts)
	} else {
		// sqlite
		db, err = gorm.Open(sqlite.Open(s.Database), opts)
	}
	if err != nil {
		log.Fatal().Err(err).Str("database", s.Database).
			Msg("couldn't open database.")
		return
	}

	// migration
	db.AutoMigrate(&User{}, &Wallet{}, &Payment{}, &BalanceCheck{}, &AppDataItem{})

	// start lightning backend
	switch s.LightningBackend {
	case "lndrest":
	case "lndgrpc":
	case "eclair":
	case "clightning":
	case "sparko":
		ln = sparko.Start(sparko.Params{
			Host:               s.SparkoURL,
			Key:                s.SparkoToken,
			InvoiceLabelPrefix: "lbs",
		})
	case "lnbits":
	default:
		// use void wallet that does nothing
		ln = void.Start()
	}
	if info, err := ln.GetInfo(); err != nil {
		log.Fatal().Err(err).Str("lightning", s.LightningBackend).
			Msg("couldn't start lightning backend.")
		return
	} else {
		log.Info().Int64("msat", info.Balance).Str("kind", s.LightningBackend).
			Msg("initialized lightning backend")
	}

	// serve http routes
	router.Path("/v/settings").HandlerFunc(viewSettings)
	router.Path("/v/lnurlscan/{code}").HandlerFunc(viewLnurlScan)
	router.Path("/v/sse").HandlerFunc(viewSSE)
	router.Path("/api/user").HandlerFunc(apiUser)
	router.Path("/api/create-wallet").HandlerFunc(apiCreateWallet)
	router.Path("/api/wallet").HandlerFunc(apiWallet)
	router.Path("/api/wallet/rename/{new-name}").HandlerFunc(apiRenameWallet)
	router.Path("/api/wallet/create-invoice").HandlerFunc(apiCreateInvoice)
	router.Path("/api/wallet/pay-invoice").HandlerFunc(apiPayInvoice)
	router.Path("/api/wallet/lnurlauth").HandlerFunc(apiLnurlAuth)
	router.Path("/api/wallet/pay-lnurl").HandlerFunc(apiPayLnurl)
	router.Path("/api/wallet/payment/{id}").HandlerFunc(apiGetPayment)

	// lnbits compatibility routes (for lnbits libraries and lnbits wallets)
	router.Path("/api/v1/wallet").HandlerFunc(apiWallet)
	router.Path("/api/v1/wallet/{new-name}").HandlerFunc(apiRenameWallet)
	router.Path("/api/v1/payments").MatcherFunc(
		func(r *http.Request, rm *mux.RouteMatch) bool {
			var outer struct {
				Out bool `json:"out"`
			}
			json.NewDecoder(r.Clone(r.Context()).Body).Decode(&outer)
			return !outer.Out
		},
	).HandlerFunc(apiCreateInvoice)
	router.Path("/api/v1/payments").MatcherFunc(
		func(r *http.Request, rm *mux.RouteMatch) bool {
			var outer struct {
				Out bool `json:"out"`
			}
			json.NewDecoder(r.Clone(r.Context()).Body).Decode(&outer)
			return outer.Out
		},
	).HandlerFunc(apiPayInvoice)
	router.Path("/api/v1/payments/lnurl").HandlerFunc(apiPayLnurl)
	router.Path("/api/v1/payments/{id}").HandlerFunc(apiGetPayment)
	router.Path("/api/v1/payments/sse").HandlerFunc(viewSSE)

	router.Use(userMiddleware)
	router.Use(walletMiddleware)

	// serve static client
	if staticFS, err := fs.Sub(static, "client/dist/spa"); err != nil {
		log.Fatal().Err(err).Msg("failed to load static files subdir")
		return
	} else {
		spaFS := SpaFS{staticFS}
		httpFS := http.FS(spaFS)
		router.PathPrefix("/").Handler(http.FileServer(httpFS))
	}

	// start http server
	log.Info().Str("host", s.Host+":"+s.Port).Msg("http listening")
	srv := &http.Server{
		Handler:      cors.Default().Handler(router),
		Addr:         s.Host + ":" + s.Port,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("error serving http")
	}
}

type SpaFS struct {
	base fs.FS
}

func (s SpaFS) Open(name string) (fs.File, error) {
	if file, err := s.base.Open(name); err == nil {
		return file, nil
	} else {
		return s.base.Open("index.html")
	}
}

type JSONError struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func jsonError(w http.ResponseWriter, code int, msg string, args ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(JSONError{false, fmt.Sprintf(msg, args...)})
	http.Error(w, string(b), code)
}
