package main

import (
	"time"

	decodepay "github.com/fiatjaf/ln-decodepay"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

func routines() {
	time.Sleep(15 * time.Minute)

	for {
		deleteExpiredInvoices()

		time.Sleep(12 * time.Hour)
	}
}

func deleteExpiredInvoices() {
	log.Info().Msg("deleting expired invoices")

	var payments []models.Payment

	storage.DB.
		Where("pending AND amount > 0 AND created_at < ?",
			time.Now().AddDate(0, 0, -8)).
		Find(&payments)

	expiredThreshold := time.Now().AddDate(0, 0, -7)

	for _, payment := range payments {
		if inv, err := decodepay.Decodepay(payment.Bolt11); err == nil {
			if time.Unix(int64(inv.CreatedAt+inv.Expiry), 0).Before(expiredThreshold) {
				storage.DB.Delete(&payment)
				log.Info().Interface("payment", payment).Msg("deleted")
			}
		}
	}
}
