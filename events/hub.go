package events

import (
	"os"

	"github.com/fiatjaf/relampago"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	"github.com/rs/zerolog"
)

type GenericEvent struct {
	App    string
	Wallet string
	Name   string
	Data   interface{}
}

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
	genericEvent    []chan GenericEvent
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

func OnGenericEvent(c chan GenericEvent) {
	subs.genericEvent = append(subs.genericEvent, c)
}

func NotifyInvoicePaid(status relampago.InvoiceStatus) {
	if !status.Paid {
		return
	}

	log := log.With().Interface("status", status).Logger()

	payment := models.Payment{}
	result := storage.DB.Where("checking_id = ?", status.CheckingID).First(&payment)
	if payment.CheckingID == "" {
		log.Warn().Err(result.Error).
			Msg("invoice paid, but no associated payment found on db?")
		return
	}

	log = log.With().Interface("payment", payment).Logger()

	result = storage.DB.
		Model(&models.Payment{}).
		Where("checking_id = ?", status.CheckingID).
		Where("amount > 0"). // means this is the receiver side of a payment, just in case
		Where("pending").
		Updates(map[string]interface{}{
			"pending": false,
			"amount":  status.MSatoshiReceived,
		})
	if result.Error != nil {
		log.Warn().Err(result.Error).Interface("payment", payment).
			Msg("failed to update payment received")
	}

	EmitPaymentReceived(payment)
}

func NotifyPaymentSentStatus(status relampago.PaymentStatus) {
	if status.Status != relampago.Failed && status.Status != relampago.Complete {
		return
	}

	log := log.With().Interface("status", status).Logger()

	payment := models.Payment{}
	result := storage.DB.Where("checking_id = ?", status.CheckingID).First(&payment)
	if payment.CheckingID == "" {
		log.Warn().Err(result.Error).
			Msg("payment sent event, but no associated payment found on db?")
		return
	}

	log = log.With().Interface("payment", payment).Logger()

	switch status.Status {
	case relampago.Failed:
		result := storage.DB.
			Model(&payment).
			Where("checking_id = ?", status.CheckingID).
			Where("pending").
			Delete(&models.Payment{})

		if result.Error != nil {
			log.Warn().Err(result.Error).
				Msg("failed to delete failed sent payment")
			return
		}

		EmitPaymentFailed(payment)

		return

	case relampago.Complete:
		result := storage.DB.
			Model(&payment).
			Where("checking_id = ?", status.CheckingID).
			Where("amount < 0"). // means this is the sender side of a payment, just in case
			Where("pending").
			Updates(map[string]interface{}{
				"pending":  false,
				"preimage": status.Preimage,
				"fee":      status.FeePaid,
			})

		if result.Error != nil {
			log.Warn().Err(result.Error).
				Msg("failed to update payment successfully sent sent")
		}

		EmitPaymentSent(payment)
	}
}

func EmitPaymentSent(payment models.Payment) {
	for _, c := range subs.paymentSent {
		c <- payment
	}
}

func EmitPaymentReceived(payment models.Payment) {
	for _, c := range subs.paymentReceived {
		c <- payment
	}
}

func EmitPaymentFailed(payment models.Payment) {
	for _, c := range subs.paymentFailed {
		c <- payment
	}
}

func EmitGenericEvent(name string, data interface{}) {
	for _, c := range subs.genericEvent {
		c <- GenericEvent{
			Name: name,
			Data: data,
		}
	}
}

func EmitGenericAppWalletEvent(app, wallet, name string, data interface{}) {
	for _, c := range subs.genericEvent {
		c <- GenericEvent{
			App:    app,
			Wallet: wallet,
			Name:   name,
			Data:   data,
		}
	}
}
