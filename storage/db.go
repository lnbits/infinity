package storage

import (
	"strings"

	"github.com/lnbits/lnbits/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(databaseConnectionString string) error {
	var err error
	opts := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}
	if strings.HasPrefix(databaseConnectionString, "postgres") {
		// postgres
		DB, err = gorm.Open(postgres.Open(databaseConnectionString), opts)
	} else if strings.HasPrefix(databaseConnectionString, "cockroach") {
		// cockroach
		DB, err = gorm.Open(postgres.Open(databaseConnectionString), opts)
	} else {
		// sqlite
		DB, err = gorm.Open(sqlite.Open(databaseConnectionString), opts)
	}
	if err != nil {
		return err
	}

	// migration
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.UserApp{},
		&models.Payment{},
		&models.BalanceCheck{},
		&models.AppDataItem{},
	); err != nil {
		return err
	}

	return nil
}
