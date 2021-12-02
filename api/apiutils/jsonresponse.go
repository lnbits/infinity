package apiutils

import (
	"net/http"

	"github.com/lnbits/lnbits/utils"
)

func SendJSON(w http.ResponseWriter, value interface{}) error {
	jsonb, err := utils.JSONMarshal(value)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonb)
	return err
}
