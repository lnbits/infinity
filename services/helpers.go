package services

import (
	"encoding/json"

	"github.com/lnbits/infinity/utils"
)

func mapToStruct(given map[string]interface{}, desiredStruct interface{}) {
	j, _ := utils.JSONMarshal(given)
	json.Unmarshal(j, &desiredStruct)
}
