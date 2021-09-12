package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fiatjaf/relampago"
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
	ThemeOptions      []string `envconfig:"LNBITS_THEME_OPTIONS" default:"classic, flamingo, mint, salvador, monochrome, autumn"`
	DefaultWalletName string   `envconfig:"LNBITS_DEFAULT_WALLET_NAME" default:"LNbits Wallet"`

	LightningBackend string `envconfig:"LNBITS_LIGHTNING_BACKEND" default:"void"`
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
	case "lnbits":
	default:
		// use void wallet that does nothing
		ln = void.New()
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
	router.Path("/api/settings").HandlerFunc(apiSettings)
	router.Path("/api/user").HandlerFunc(apiUser)
	router.Path("/api/wallet").HandlerFunc(apiWallet)

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
