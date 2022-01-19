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
	"github.com/lnbits/lnbits/nostr"
	"github.com/lnbits/lnbits/services"
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
	Secret   string `envconfig:"SECRET" required`

	SiteTitle         string   `envconfig:"LNBITS_SITE_TITLE" default:"LNBitsLocal"`
	SiteTagline       string   `envconfig:"LNBITS_SITE_TAGLINE" default:"Locally-hosted lightning wallet"`
	SiteDescription   string   `envconfig:"LNBITS_SITE_DESCRIPTION" default:""`
	DefaultWalletName string   `envconfig:"LNBITS_DEFAULT_WALLET_NAME" default:"LNbits Wallet"`
	AppCacheSize      int      `envconfig:"APP_CACHE_SIZE" default:"200"`
	NostrRelays       []string `envconfig:"NOSTR_RELAYS"`
	TunnelDomain      string   `envconfig:"TUNNEL_DOMAIN"`

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
	services.Secret = s.Secret
	nostr.Relays = s.NostrRelays
	nostr.Secret = s.Secret

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

	// start nostr
	nostr.Start()

	// start routines
	go routines()

	// serve tunnel clients
	if s.TunnelDomain != "" {
		services.TunnelDomain = s.TunnelDomain
		router.Host("{subdomain:[a-z0-9]+}." + s.TunnelDomain).
			Handler(services.TunnelHandler)
	}

	// serve http routes
	mainroute := router.NewRoute()
	if s.ServerName != "" {
		mainroute = mainroute.Host(s.ServerName)
	}

	mainroute.Path("/v/settings").HandlerFunc(viewSettings)
	mainroute.Path("/api/user").HandlerFunc(api.User)
	mainroute.Path("/api/user/create-wallet").HandlerFunc(api.CreateWallet)
	mainroute.Path("/api/user/add-app").HandlerFunc(api.AddApp)
	mainroute.Path("/api/user/remove-app").HandlerFunc(api.RemoveApp)
	mainroute.Path("/api/wallet").HandlerFunc(api.Wallet)
	mainroute.Path("/api/wallet/delete").HandlerFunc(api.DeleteWallet)
	mainroute.Path("/api/wallet/rename/{new-name}").HandlerFunc(api.RenameWallet)
	mainroute.Path("/api/wallet/create-invoice").HandlerFunc(api.CreateInvoice)
	mainroute.Path("/api/wallet/pay-invoice").HandlerFunc(api.PayInvoice)
	mainroute.Path("/api/wallet/lnurlauth").HandlerFunc(api.LnurlAuth)
	mainroute.Path("/api/wallet/pay-lnurl").HandlerFunc(api.PayLnurl)
	mainroute.Path("/api/wallet/payment/{id}").HandlerFunc(api.GetPayment)
	mainroute.Path("/api/wallet/lnurlscan/{code}").HandlerFunc(api.LnurlScan)
	mainroute.Path("/api/wallet/sse").HandlerFunc(api.SSE)
	mainroute.Path("/lnurl/wallet/drain").HandlerFunc(api.DrainFunds)

	// app endpoints
	mainroute.Path("/api/wallet/app/sse").HandlerFunc(apps.SSE)
	mainroute.Path("/api/wallet/app/{appid}").HandlerFunc(apps.Info)
	mainroute.Path("/api/wallet/app/{appid}/refresh").HandlerFunc(apps.Refresh)
	mainroute.Path("/api/wallet/app/{appid}/list/{model}").HandlerFunc(apps.ListItems)
	mainroute.Path("/api/wallet/app/{appid}/get/{model}/{key}").HandlerFunc(apps.GetItem)
	mainroute.Path("/api/wallet/app/{appid}/set/{model}/{key}").HandlerFunc(apps.SetItem)
	mainroute.Path("/api/wallet/app/{appid}/add/{model}").HandlerFunc(apps.AddItem)
	mainroute.Path("/api/wallet/app/{appid}/del/{model}/{key}").HandlerFunc(apps.DeleteItem)
	mainroute.Path("/ext/{wallet}/{appid}/action/{action}").HandlerFunc(apps.CustomAction)
	mainroute.Path("/ext/{wallet}/{appid}/sse").HandlerFunc(apps.PublicSSE)
	mainroute.PathPrefix("/ext/{wallet}/{appid}/").HandlerFunc(apps.StaticFile)

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
	services.MainServer = srv
	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("error serving http")
	}
}
