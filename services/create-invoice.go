package services

import (
	"encoding/hex"
	"fmt"
	"time"

	decodepay "github.com/fiatjaf/ln-decodepay"
	"github.com/lnbits/lnbits/lightning"
	"github.com/lnbits/lnbits/models"
	"github.com/lnbits/lnbits/storage"
	rp "github.com/lnbits/relampago"
)

var DefaultInvoiceExpiry = time.Minute * 15

type CreateInvoiceParams struct {
	rp.InvoiceParams

	Tag     string            `json:"tag"`
	Extra   models.JSONObject `json:"extra"`
	Webhook string            `json:"webhook"`
}

func CreateInvoiceFromApp(walletID string, params map[string]interface{}) (interface{}, error) {
	var s CreateInvoiceParams
	mapToStruct(params, &s)
	if h, ok := params["description_hash"]; ok {
		s.DescriptionHash, _ = hex.DecodeString(h.(string))
	}
	if m, ok := params["msatoshi"]; ok {
		if f, ok := m.(float64); ok {
			s.Msatoshi = int64(f)
		}
	}
	return CreateInvoice(walletID, s)
}

func CreateInvoice(walletID string, params CreateInvoiceParams) (models.Payment, error) {
	params.InvoiceParams.Expiry = &DefaultInvoiceExpiry

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
