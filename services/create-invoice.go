package services

import (
	"encoding/json"
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
	j, _ := json.Marshal(params)
	var s CreateInvoiceParams
	json.Unmarshal(j, &s)
	payment, err := CreateInvoice(walletID, s)
	if err != nil {
		return nil, err
	}
	j, _ = json.Marshal(payment)
	var resp interface{}
	json.Unmarshal(j, &resp)
	return resp, nil
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
	}
	if result := storage.DB.Create(&payment); result.Error != nil {
		return payment, fmt.Errorf("failed to save invoice: %w", result.Error)
	}

	return payment, nil
}
