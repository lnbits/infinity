package services

import (
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"gorm.io/gorm"
)

func LoadWalletPayments(walletID string) ([]models.Payment, error) {
	var payments []models.Payment

	result := storage.DB.
		Order("created_at desc").
		Where("wallet_id = ?", walletID).
		Find(&payments)

	if result.Error == gorm.ErrRecordNotFound {
		return payments, nil
	}

	return payments, result.Error
}
