package services

import (
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func GetWalletPaymentFromApp(walletID string, id string) (interface{}, error) {
	payment, err := GetWalletPayment(walletID, id)
	if err != nil {
		return nil, err
	}
	return structToInterface(payment), nil
}

func GetWalletPayment(walletID string, hashOrCheckingID string) (models.Payment, error) {
	var payment models.Payment

	result := storage.DB.
		Where("wallet_id = ?", walletID).
		Where("hash = ?", hashOrCheckingID).Or("checking_id = ?", hashOrCheckingID).
		First(&payment)

	return payment, result.Error
}
