package services

import (
	"encoding/json"

	"github.com/lnbits/lnbits/utils"
)

func mapToStruct(given map[string]interface{}, desiredStruct interface{}) {
	j, _ := utils.JSONMarshal(given)
	json.Unmarshal(j, &desiredStruct)
}
