module github.com/lnbits/lnbits

go 1.16

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-sqlite3 v1.14.8 // indirect
	github.com/rs/zerolog v1.25.0
	gorm.io/driver/postgres v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.15
)

replace github.com/fiatjaf/relampago => /home/fiatjaf/comp/relampago
