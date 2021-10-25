package services

import (
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func LoadWalletBalance(walletID string) (int64, error) {
	var balance int64

	result := storage.DB.Model(&models.Payment{}).
		Select("coalesce(sum(amount), 0)").
		Where("amount < 0 OR (amount > 0 AND NOT pending)").
		Where("wallet_id = ?", walletID).
		First(&balance)

	return balance, result.Error
}
