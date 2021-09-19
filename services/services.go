package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fiatjaf/relampago"
	"github.com/lnbits/lnbits/lightning"
	"github.com/lnbits/lnbits/storage"
	"gorm.io/gorm"
)

var ln relampago.Wallet
var db *gorm.DB
var httpClient = &http.Client{
	Timeout: time.Second * 7,
	CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return fmt.Errorf("target '%s' has returned a redirect", r.URL)
	},
}

func init() {
	ln = lightning.LN
	db = storage.DB
}
