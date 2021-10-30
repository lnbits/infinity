package services

import (
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func LoadWalletPayments(walletID string) ([]models.Payment, error) {
	var payments []models.Payment

	result := storage.DB.
		Order("created_at desc").
		Where("wallet_id = ?", walletID).
		Find(&payments)

	return payments, result.Error
}
