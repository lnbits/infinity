package main

import (
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/lnbits/lnbits/api"
	"github.com/lnbits/lnbits/apps"
	"github.com/lnbits/lnbits/lightning"
	"github.com/lnbits/lnbits/storage"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

type Settings struct {
	Host            string   `envconfig:"HOST" default:"0.0.0.0"`
	Port            string   `envconfig:"PORT" default:"5000"`
	QuasarDevServer *url.URL `envconfig:"QUASAR_DEV_SERVER"`
	ServerName      string   `envconfig:"SERVER_NAME"`

	Database string `envconfig:"DATABASE" default:"dev.sqlite"`

	SiteTitle         string `envconfig:"LNBITS_SITE_TITLE" default:"LNBitsLocal"`
	SiteTagline       string `envconfig:"LNBITS_SITE_TAGLINE" default:"Locally-hosted lightning wallet"`
	SiteDescription   string `envconfig:"LNBITS_SITE_DESCRIPTION" default:""`
	DefaultWalletName string `envconfig:"LNBITS_DEFAULT_WALLET_NAME" default:"LNbits Wallet"`

	AppCacheSize int `envconfig:"APP_CACHE_SIZE" default:"128"`

	LightningBackend string `envconfig:"LNBITS_LIGHTNING_BACKEND" default:"void"`
	// -- other env vars are defined in the 'lightning' package
}

var s Settings
var log = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stdout})
var router = mux.NewRouter()
var commit string // will be set at compile time

func main() {
	// environment variables
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig.")
		return
	}
	apps.AppCacheSize = s.AppCacheSize
	apps.ServerName = s.ServerName
	api.SiteTitle = s.SiteTitle

	// setup logger
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log = log.With().Timestamp().Logger()
	apps.SetLogger(log)

	// database
	if err := storage.Connect(s.Database); err != nil {
		log.Fatal().Err(err).Str("database", s.Database).
			Msg("couldn't open database.")
	}

	// lightning backend
	go func() {
		lightning.Connect(s.LightningBackend)
		if info, err := lightning.LN.GetInfo(); err != nil {
			log.Fatal().Err(err).Str("lightning", s.LightningBackend).
				Msg("couldn't start lightning backend.")
			return
		} else {
			log.Info().Int64("msat", info.Balance).Str("kind", s.LightningBackend).
				Msg("initialized lightning backend")
		}
	}()

	// start routines
	go routines()

	// serve http routes
	router.Path("/v/settings").HandlerFunc(viewSettings)
	router.Path("/api/user").HandlerFunc(api.User)
	router.Path("/api/user/create-wallet").HandlerFunc(api.CreateWallet)
	router.Path("/api/user/add-app").HandlerFunc(api.AddApp)
	router.Path("/api/user/remove-app").HandlerFunc(api.RemoveApp)
	router.Path("/api/wallet").HandlerFunc(api.Wallet)
	router.Path("/api/wallet/delete").HandlerFunc(api.DeleteWallet)
	router.Path("/api/wallet/rename/{new-name}").HandlerFunc(api.RenameWallet)
	router.Path("/api/wallet/create-invoice").HandlerFunc(api.CreateInvoice)
	router.Path("/api/wallet/pay-invoice").HandlerFunc(api.PayInvoice)
	router.Path("/api/wallet/lnurlauth").HandlerFunc(api.LnurlAuth)
	router.Path("/api/wallet/pay-lnurl").HandlerFunc(api.PayLnurl)
	router.Path("/api/wallet/payment/{id}").HandlerFunc(api.GetPayment)
	router.Path("/api/wallet/lnurlscan/{code}").HandlerFunc(api.LnurlScan)
	router.Path("/api/wallet/sse").HandlerFunc(api.SSE)
	router.Path("/lnurl/wallet/drain").HandlerFunc(api.DrainFunds)

	// app endpoints
	router.Path("/api/wallet/app/sse").HandlerFunc(apps.SSE)
	router.Path("/api/wallet/app/{appid}").HandlerFunc(apps.Info)
	router.Path("/api/wallet/app/{appid}/refresh").HandlerFunc(apps.Refresh)
	router.Path("/api/wallet/app/{appid}/list/{model}").HandlerFunc(apps.ListItems)
	router.Path("/api/wallet/app/{appid}/get/{model}/{key}").HandlerFunc(apps.GetItem)
	router.Path("/api/wallet/app/{appid}/set/{model}/{key}").HandlerFunc(apps.SetItem)
	router.Path("/api/wallet/app/{appid}/add/{model}").HandlerFunc(apps.AddItem)
	router.Path("/api/wallet/app/{appid}/del/{model}/{key}").HandlerFunc(apps.DeleteItem)
	router.Path("/ext/{wallet}/{appid}/action/{action}").HandlerFunc(apps.CustomAction)
	router.Path("/ext/{wallet}/{appid}/sse").HandlerFunc(apps.PublicSSE)
	router.PathPrefix("/ext/{wallet}/{appid}/").HandlerFunc(apps.StaticFile)

	// middleware
	router.Use(handlers.ProxyHeaders)
	router.Use(jsonHeaderMiddleware)
	router.Use(userMiddleware)
	router.Use(walletMiddleware)
	router.Use(cors.AllowAll().Handler)

	serveStaticClient()

	// start http server
	log.Info().Str("host", s.Host+":"+s.Port).Msg("http listening")
	srv := &http.Server{
		Handler:      router,
		Addr:         s.Host + ":" + s.Port,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("error serving http")
	}
}
