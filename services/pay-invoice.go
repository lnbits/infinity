package services

import (
	"fmt"
	"strings"

	decodepay "github.com/fiatjaf/ln-decodepay"
	"github.com/lnbits/infinity/events"
	"github.com/lnbits/infinity/lightning"
	"github.com/lnbits/infinity/models"
	"github.com/lnbits/infinity/storage"
	"github.com/lnbits/infinity/utils"
	rp "github.com/lnbits/relampago"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PayInvoiceParams struct {
	rp.PaymentParams

	Tag     string            `json:"tag"`
	Extra   models.JSONObject `json:"extra"`
	Webhook string            `json:"webhook"`
}

func PayInvoiceFromApp(walletID string, params map[string]interface{}) (interface{}, error) {
	var s PayInvoiceParams
	mapToStruct(params, &s)
	return PayInvoice(walletID, s)
}

func PayInvoice(walletID string, params PayInvoiceParams) (payment models.Payment, err error) {
	// parse invoice
	inv, err := decodepay.Decodepay(params.Invoice)
	if err != nil {
		return payment, fmt.Errorf("failed to parse invoice: %w", err)
	}

	// get amount we will pay
	var invoiceAmount int64
	if params.CustomAmount != 0 {
		if params.CustomAmount < inv.MSatoshi {
			return payment, fmt.Errorf(
				"custom amount %d is smaller than invoice amount %d",
				params.CustomAmount, inv.MSatoshi)
		}

		invoiceAmount = params.CustomAmount
	} else {
		invoiceAmount = inv.MSatoshi
	}

	// add payment to database first
	temp := "tmp_" + utils.RandomHex(16)
	payment = models.Payment{
		CheckingID:  temp,
		Pending:     true,
		Amount:      -invoiceAmount,
		Hash:        inv.PaymentHash,
		Bolt11:      params.Invoice,
		Tag:         params.Tag,
		Extra:       params.Extra,
		Webhook:     params.Webhook,
		WalletID:    walletID,
		Description: inv.Description,
		Fee:         invoiceAmount / 100,
	}
	if result := storage.DB.Create(&payment); result.Error != nil {
		return payment, fmt.Errorf("failed to save temp payment: %w", result.Error)
	}

	defer func() {
		if err != nil {
			result := storage.DB.Where("checking_id", temp).Delete(&payment)
			if result.Error != nil {
				panic("failed to delete temp payment " + payment.CheckingID + ": " +
					result.Error.Error())
			}
		}
	}()

	// check balance
	if balance, err := LoadWalletBalance(walletID); err != nil {
		return payment, fmt.Errorf("failed to check balance: %w", err)
	} else if balance <= 0 {
		return payment, fmt.Errorf("insufficient balance: needs %d more msat", -balance)
	}

	// check if this is an internal payment
	var internal models.Payment
	result := storage.DB.
		Where("hash = ?", payment.Hash).
		Where("amount > 0").
		Where("pending"). // search for a pending incoming payment with this same hash
		First(&internal)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return payment, fmt.Errorf("failed to check internal payment: %w", result.Error)
	}
	if internal.CheckingID != "" {
		// if internal, settle it
		newSenderCheckingID := strings.Replace(payment.CheckingID, "tmp_", "int_", 1)

		go func() {
			err := storage.DB.Transaction(func(tx *gorm.DB) error {
				result := tx.Model(&models.Payment{}).
					Where("checking_id", internal.CheckingID).
					Update("pending", false)
				if result.Error != nil {
					return result.Error
				}

				result = tx.Model(&models.Payment{}).
					Where("checking_id", payment.CheckingID).
					Updates(map[string]interface{}{
						"checking_id": newSenderCheckingID,
						"pending":     false,
					})
				if result.Error != nil {
					return result.Error
				}

				return nil
			})
			if err != nil {
				log.Error().Err(err).Str("receiving", internal.CheckingID).
					Str("paying", payment.CheckingID).
					Msg("failed to settle internal payment")
				return
			}

			// internal settlement has succeeded, emit events
			payment.CheckingID = newSenderCheckingID
			payment.Pending = false
			events.EmitPaymentSent(payment)

			internal.Pending = false
			events.EmitPaymentReceived(internal)
		}()
	}

	// actually perform the payment
	data, err := lightning.LN.MakePayment(params.PaymentParams)
	if err != nil {
		return payment, fmt.Errorf("failed to pay: %w", err)
	}

	// update checking_id
	result = storage.DB.
		Model(&models.Payment{}).
		Where("checking_id", temp).
		Update("checking_id", data.CheckingID)
	if result.Error != nil {
		return payment, fmt.Errorf("failed to update checking_id: %w", result.Error)
	}

	return payment, nil
}
