package services

import (
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
	"github.com/lnbits/infinity/utils"
	"gorm.io/gorm"
)

func Transfer(walletID string, toWalletID string, msatoshi int64, desc string) error {
	sharedHash := utils.RandomHex(16)

	leaving := models.Payment{
		Hash:        sharedHash,
		CheckingID:  "int_" + utils.RandomHex(32),
		Amount:      -msatoshi,
		Description: desc,
		Tag:         "transfer",
		WalletID:    walletID,
	}

	entering := models.Payment{
		Hash:        sharedHash,
		CheckingID:  "int_" + utils.RandomHex(32),
		Amount:      msatoshi,
		Description: desc,
		Tag:         "transfer",
		WalletID:    toWalletID,
	}

	err := storage.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(leaving).Error; err != nil {
			return err
		}
		if err := tx.Create(entering).Error; err != nil {
			return err
		}
		return nil
	})

	return err
}
