module github.com/lnbits/lnbits

go 1.16

require (
	github.com/aarzilli/golua v0.0.0-20210507130708-11106aa57765
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/btcsuite/btcd v0.20.1-beta.0.20200515232429-9f0179fd2c46
	github.com/fiatjaf/go-lnurl v1.6.1
	github.com/fiatjaf/ln-decodepay v1.1.0
	github.com/fiatjaf/lunatico v1.0.0
	github.com/fiatjaf/relampago v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pretty v0.3.0 // indirect
	github.com/lucsky/cuid v1.2.1
	github.com/mattn/go-sqlite3 v1.14.8 // indirect
	github.com/rs/cors v1.8.0
	github.com/rs/zerolog v1.25.0
	github.com/tidwall/gjson v1.9.0
	gopkg.in/antage/eventsource.v1 v1.0.0-20150318155416-803f4c5af225
	gorm.io/driver/postgres v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.15
)

replace github.com/fiatjaf/relampago => /home/fiatjaf/comp/relampago

replace github.com/fiatjaf/go-lnurl => /home/fiatjaf/comp/go-lnurl

replace github.com/fiatjaf/lunatico => /home/fiatjaf/comp/lunatico
