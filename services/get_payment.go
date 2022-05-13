package services

import (
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
)

func GetWalletPayment(walletID string, hashOrCheckingID string) (models.Payment, error) {
	var payment models.Payment

	result := storage.DB.
		Where("wallet_id = ?", walletID).
		Where("hash = ?", hashOrCheckingID).Or("checking_id = ?", hashOrCheckingID).
		First(&payment)

	return payment, result.Error
}
