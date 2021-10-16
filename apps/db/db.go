package db

import (
	"fmt"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/lucsky/cuid"
	"gorm.io/gorm/clause"
)

func Get(wallet string, app string, key string) (map[string]interface{}, error) {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Key:      key,
	}

	result := storage.DB.Where(&item).First(&item)
	return item.Value, result.Error
}

func Set(wallet string, app string, key string, value map[string]interface{}) error {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Key:      key,
		Value:    value,
	}

	result := storage.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app"}, {Name: "wallet_id"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&item)

	return result.Error
}

func Add(wallet string, app string, value map[string]interface{}) error {
	return Set(wallet, app, cuid.Slug(), value)
}

func Update(wallet string, app string, key string, updates map[string]interface{}) error {
	value, err := Get(wallet, app, key)
	if err != nil {
		return fmt.Errorf("failed to get %s: %w", key, err)
	}

	for k, v := range updates {
		value[k] = v
	}

	return Set(wallet, app, key, value)
}

func Delete(wallet string, app string, key string) error {
	result := storage.DB.Delete(&models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Key:      key,
	})

	return result.Error
}
