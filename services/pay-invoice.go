package services

import (
	"fmt"

	decodepay "github.com/fiatjaf/ln-decodepay"
	rp "github.com/fiatjaf/relampago"
	m "github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/utils"
)

type PayInvoiceParams struct {
	rp.PaymentParams

	Tag     string       `json:"tag"`
	Extra   m.JSONObject `json:"extra"`
	Webhook string       `json:"webhook"`
}

func PayInvoice(wallet *m.Wallet, params PayInvoiceParams) (payment m.Payment, err error) {
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
	temp := "tmp" + utils.RandomHex(16)
	payment = m.Payment{
		CheckingID: temp,
		Pending:    true,
		Amount:     -invoiceAmount,
		Hash:       inv.PaymentHash,
		Bolt11:     params.Invoice,
		Tag:        params.Tag,
		Extra:      params.Extra,
		Webhook:    params.Webhook,
		WalletID:   wallet.ID,
	}
	if result := db.Create(&payment); result.Error != nil {
		return payment, fmt.Errorf("failed to save temp payment: %w", result.Error)
	}

	defer func() {
		if err != nil {
			if result := db.Delete(&payment); result.Error != nil {
				panic("failed to delete temp payment " + payment.CheckingID + ": " +
					result.Error.Error())
			}
		}
	}()

	// check balance
	var balance int64
	if result := db.Model(&m.Payment{}).
		Select("sum(amount)").
		Where("amount < 0 OR (amount > 0 AND NOT pending)").
		Where("wallet_id = ?", wallet.ID).
		First(&balance); result.Error != nil {
		return payment, fmt.Errorf("failed to check balance: %w", result.Error)
	}

	if balance <= 0 {
		return payment, fmt.Errorf("insufficient balance: needs %d more msat", -balance)
	}

	// actually perform the payment
	data, err := ln.MakePayment(params.PaymentParams)
	if err != nil {
		return payment, fmt.Errorf("failed to pay: %w", err)
	}

	// update checking_id
	payment.CheckingID = data.CheckingID
	if result := db.Save(payment); result.Error != nil {
		return payment, fmt.Errorf("failed to update checking_id: %w", err)
	}

	return payment, nil
}
