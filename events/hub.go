package events

import (
	"os"

	"github.com/fiatjaf/relampago"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/rs/zerolog"
)

var log = zerolog.
	New(os.Stderr).
	Output(zerolog.ConsoleWriter{Out: os.Stdout}).
	With().
	Str("s", "events").
	Logger()

var subs = struct {
	paymentReceived []chan models.Payment
	paymentSent     []chan models.Payment
	paymentFailed   []chan models.Payment
}{}

func OnPaymentReceived(c chan models.Payment) {
	subs.paymentReceived = append(subs.paymentReceived, c)
}

func OnPaymentSent(c chan models.Payment) {
	subs.paymentSent = append(subs.paymentSent, c)
}

func OnPaymentFailed(c chan models.Payment) {
	subs.paymentFailed = append(subs.paymentFailed, c)
}

func EmitInvoicePaid(status relampago.InvoiceStatus) {
	if !status.Paid {
		return
	}

	payment := models.Payment{}
	result := storage.DB.
		Where("checking_id = ?", status.CheckingID).
		Where("amount > 0"). // means this is the receiver side of a payment, just in case
		Where("pending").
		Updates(map[string]interface{}{
			"pending": false,
			"amount":  status.MSatoshiReceived,
		}).
		First(&payment)
	if result.Error != nil {
		log.Warn().Err(result.Error).
			Msg("invoice paid, but no associated payment found on db?")
	}

	for _, c := range subs.paymentReceived {
		c <- payment
	}
}

func EmitPaymentSent(status relampago.PaymentStatus) {
	payment := models.Payment{}

	switch status.Status {
	case relampago.Failed:
		result := storage.DB.
			Where("checking_id = ?", status.CheckingID).
			Where("pending").
			Delete(&models.Payment{}).
			Find(&payment)

		if result.Error != nil {
			log.Warn().Err(result.Error).Str("checking_id", status.CheckingID).
				Msg("failed to delete failed sent payment")
			return
		}

		for _, c := range subs.paymentFailed {
			c <- payment
		}

		return

	case relampago.Complete:
		result := storage.DB.
			Where("checking_id = ?", status.CheckingID).
			Where("amount < 0"). // means this is the sender side of a payment, just in case
			Where("pending").
			Updates(map[string]interface{}{
				"pending":  false,
				"preimage": status.Preimage,
				"fee":      status.FeePaid,
			}).
			First(&payment)

		if result.Error != nil {
			log.Warn().Err(result.Error).
				Msg("payment sent, but no associated payment found on db?")
		}

		for _, c := range subs.paymentSent {
			c <- payment
		}
	}
}
