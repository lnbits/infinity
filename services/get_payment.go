package services

import (
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func GetWalletPayment(walletID string, hashOrCheckingID string) (models.Payment, error) {
	var payment models.Payment

	result := storage.DB.
		Where("wallet_id = ?", walletID).
		Where("hash = ?", hashOrCheckingID).Or("checking_id = ?", hashOrCheckingID).
		First(&payment)

	return payment, result.Error
}
