package services

import (
	"fmt"

	decodepay "github.com/fiatjaf/ln-decodepay"
	rp "github.com/fiatjaf/relampago"
	"github.com/lnbits/lnbits/lightning"
	m "github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

type CreateInvoiceParams struct {
	rp.InvoiceParams

	Tag     string       `json:"tag"`
	Extra   m.JSONObject `json:"extra"`
	Webhook string       `json:"webhook"`
}

func CreateInvoice(wallet *m.Wallet, params CreateInvoiceParams) (m.Payment, error) {
	data, err := lightning.LN.CreateInvoice(params.InvoiceParams)
	if err != nil {
		return m.Payment{}, fmt.Errorf("failed to create invoice: %w", err)
	}

	inv, err := decodepay.Decodepay(data.Invoice)
	if err != nil {
		return m.Payment{}, fmt.Errorf(
			"failed to parse created invoice (%s): %w", data.Invoice, err)
	}

	payment := m.Payment{
		CheckingID: data.CheckingID,
		Pending:    true,
		Preimage:   data.Preimage,
		Hash:       inv.PaymentHash,
		Bolt11:     data.Invoice,
		Amount:     params.Msatoshi,
		WalletID:   wallet.ID,
	}
	if result := storage.DB.Create(&payment); result.Error != nil {
		return payment, fmt.Errorf("failed to save invoice: %w", result.Error)
	}

	return payment, nil
}
