package services

import (
	"fmt"

	decodepay "github.com/fiatjaf/ln-decodepay"
	rp "github.com/fiatjaf/relampago"
	"github.com/lnbits/lnbits/lightning"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
)

type CreateInvoiceParams struct {
	rp.InvoiceParams

	Tag     string            `json:"tag"`
	Extra   models.JSONObject `json:"extra"`
	Webhook string            `json:"webhook"`
}

func CreateInvoiceFromApp(walletID string, params map[string]interface{}) (interface{}, error) {
	var s CreateInvoiceParams
	mapToStruct(params, &s)
	return CreateInvoice(walletID, s)
}

func CreateInvoice(walletID string, params CreateInvoiceParams) (models.Payment, error) {
	data, err := lightning.LN.CreateInvoice(params.InvoiceParams)
	if err != nil {
		return models.Payment{}, fmt.Errorf("failed to create invoice: %w", err)
	}

	inv, err := decodepay.Decodepay(data.Invoice)
	if err != nil {
		return models.Payment{}, fmt.Errorf(
			"failed to parse created invoice (%s): %w", data.Invoice, err)
	}

	payment := models.Payment{
		CheckingID:  data.CheckingID,
		Pending:     true,
		Preimage:    data.Preimage,
		Hash:        inv.PaymentHash,
		Bolt11:      data.Invoice,
		Amount:      params.Msatoshi,
		WalletID:    walletID,
		Description: params.Description,
		Extra:       params.Extra,
		Tag:         params.Tag,
	}
	if result := storage.DB.Create(&payment); result.Error != nil {
		return payment, fmt.Errorf("failed to save invoice: %w", result.Error)
	}

	return payment, nil
}
