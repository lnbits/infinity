package api

import (
	"gorm.io/gorm"

	storage "github.com/lnbits/lnbits/storage"
)

var db *gorm.DB

func init() {
	db = storage.DB
}
