package services

import "encoding/json"

func mapToStruct(given map[string]interface{}, desiredStruct interface{}) {
	j, _ := json.Marshal(given)
	json.Unmarshal(j, &desiredStruct)
}

func structToInterface(anyStruct interface{}) interface{} {
	j, _ := json.Marshal(anyStruct)
	var result interface{}
	json.Unmarshal(j, &result)
	return result
}
