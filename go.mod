module github.com/lnbits/lnbits

go 1.16

require (
	github.com/aarzilli/golua v0.0.0-20210507130708-11106aa57765
	github.com/btcsuite/btcd v0.20.1-beta.0.20200515232429-9f0179fd2c46
	github.com/fiatjaf/go-lnurl v1.7.2
	github.com/fiatjaf/ln-decodepay v1.1.0
	github.com/fiatjaf/lunatico v1.4.0
	github.com/fiatjaf/relampago v1.0.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lucsky/cuid v1.2.1
	github.com/mattn/go-sqlite3 v1.14.8 // indirect
	github.com/rif/cache2go v1.0.0
	github.com/rs/cors v1.8.0
	github.com/rs/zerolog v1.25.0
	github.com/tidwall/gjson v1.9.0
	github.com/wI2L/jettison v0.7.3
	gopkg.in/antage/eventsource.v1 v1.0.0-20150318155416-803f4c5af225
	gorm.io/driver/postgres v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.15
)

replace github.com/fiatjaf/relampago => /home/fiatjaf/comp/relampago
