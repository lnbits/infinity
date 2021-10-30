package services

import "encoding/json"

func mapToStruct(given map[string]interface{}, desiredStruct interface{}) {
	j, _ := json.Marshal(given)
	json.Unmarshal(j, &desiredStruct)
}
