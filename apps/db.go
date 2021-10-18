package apps

import (
	"encoding/json"
	"fmt"

	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/lucsky/cuid"
	"gorm.io/gorm/clause"
)

func DBGet(wallet, app, model, key string) (map[string]interface{}, error) {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Model:    model,
		Key:      key,
	}

	result := storage.DB.Where(&item).First(&item)
	return item.Value, result.Error
}

func DBSet(wallet, app, model, key string, value map[string]interface{}) error {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Model:    model,
		Key:      key,
		Value:    value,
	}

	_, settings, err := GetAppSettings(app)
	if err != nil {
		return fmt.Errorf("failed to get app on model.set: %w", err)
	}
	if err := settings.getModel(model).validateItem(item); err != nil {
		j, _ := json.Marshal(value)
		return fmt.Errorf("invalid value %s for model %s: %w", string(j), model, err)
	}

	result := storage.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "app"}, {Name: "wallet_id"}, {Name: "model"}, {Name: "key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&item)

	if result.Error != nil {
		return result.Error
	}

	SendItemSSE(item)
	return nil
}

func DBAdd(wallet, app, model string, value map[string]interface{}) error {
	return DBSet(wallet, app, model, cuid.Slug(), value)
}

func DBUpdate(wallet, app, model, key string, updates map[string]interface{}) error {
	value, err := DBGet(wallet, app, model, key)
	if err != nil {
		return fmt.Errorf("failed to get %s: %w", key, err)
	}

	for k, v := range updates {
		value[k] = v
	}

	return DBSet(wallet, app, model, key, value)
}

func DBDelete(wallet, app, model, key string) error {
	item := models.AppDataItem{
		WalletID: wallet,
		App:      app,
		Model:    model,
		Key:      key,
	}
	result := storage.DB.Delete(&item)

	if result.Error != nil {
		return result.Error
	}

	// an item with an empty .Value means it was deleted
	SendItemSSE(item)

	return nil
}
