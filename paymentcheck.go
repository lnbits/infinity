package main

import (
	"gorm.io/gorm"

	"github.com/lnbits/infinity/events"
	"github.com/lnbits/infinity/lightning"
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
	"github.com/lnbits/relampago"
)

func initialPaymentCheck() {
	log.Info().Msg("performing startup check for pending payments and invoice")

	var payments []models.Payment
	result := storage.DB.Where("pending").Find(&payments)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Fatal().Err(result.Error).Msg("failed to query pending invoices or payments")
		return
	}

	log.Info().Msgf("will check %d payments", len(payments))
	for _, payment := range payments {
		log := log.With().Str("id", payment.CheckingID).Logger()
		log.Info().Int64("amount", payment.Amount).Msg("checking")

		if payment.Amount > 0 {
			status, err := lightning.LN.GetInvoiceStatus(payment.CheckingID)
			if err != nil {
				log.Warn().Err(err).Msg("failed to get invoice status")
				continue
			}
			if status.Paid {
				log.Info().Msg("invoice paid, updating")
				result = storage.DB.Set("pending", false).Where("checkingID", payment.CheckingID)
			} else {
				continue
			}

			if result.Error != nil {
				log.Error().Err(result.Error).Interface("status", status).Msg("failed to update invoice status")
			} else {
				events.NotifyInvoicePaid(status)
			}
		} else {
			status, err := lightning.LN.GetPaymentStatus(payment.CheckingID)
			if err != nil {
				log.Warn().Err(err).Msg("failed to get payment status")
				continue
			}
			if status.Status == relampago.Complete {
				log.Info().Str("preimage", status.Preimage).Msg("payment complete, updating")
				result = storage.DB.Set("pending", false).Where("checkingID", payment.CheckingID)
			} else if status.Status == relampago.Failed {
				log.Info().Msg("payment failed, deleting")
				result = storage.DB.Delete(&models.Payment{}).Where("checkingID", payment.CheckingID)
			} else {
				log.Info().Interface("status", status.Status).Msg("payment not complete or failed")
				continue
			}

			if result.Error != nil {
				log.Error().Err(result.Error).Interface("status", status).Msg("failed to update payment status")
			} else {
				events.NotifyPaymentSentStatus(status)
			}
		}
	}
}
