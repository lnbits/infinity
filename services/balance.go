package services

import (
	"github.com/lnbits/lnbits/storage"
	"gorm.io/gorm"
)

func LoadWalletBalance(walletID string) (int64, error) {
	var balance int64

	result := storage.DB.Raw(`
        SELECT coalesce(sum(amount), 0)
        FROM payments
        WHERE (amount < 0 OR (amount > 0 AND NOT pending))
        AND wallet_id = ?
	`, walletID).Scan(&balance)
	if result.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}

	return balance, result.Error
}
