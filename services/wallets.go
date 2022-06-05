package services

import (
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
	"github.com/lnbits/infinity/utils"
	"github.com/lucsky/cuid"
)

func CreateUser() (*models.User, error) {
	var user models.User
	user.ID = cuid.Slug()
	user.Apps = make(models.StringList, 0)
	masterKey := utils.RandomHex(32)
	user.MasterKey = masterKey
	result := storage.DB.Create(&user)
	return &user, result.Error
}

func CreateWallet(userID string, name string) (*models.Wallet, error) {
	wallet := models.Wallet{
		ID:         cuid.Slug(),
		Name:       name,
		UserID:     userID,
		InvoiceKey: utils.RandomHex(32),
		AdminKey:   utils.RandomHex(32),
	}
	result := storage.DB.Create(&wallet)
	return &wallet, result.Error
}
