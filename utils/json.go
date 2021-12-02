package utils

import (
	"encoding/json"

	"github.com/wI2L/jettison"
)

func JSONParse(jsonstr string) (interface{}, error) {
	var jsonvalue interface{}
	err := json.Unmarshal([]byte(jsonstr), &jsonvalue)
	return jsonvalue, err
}

func JSONEncode(jsonvalue interface{}) (string, error) {
	jsonb, err := JSONMarshal(jsonvalue)
	if err != nil {
		return "", err
	}
	return string(jsonb), nil
}

func JSONMarshal(value interface{}) ([]byte, error) {
	return jettison.MarshalOpts(value,
		jettison.NoHTMLEscaping(),
		jettison.UnixTime(),
		jettison.NilSliceEmpty(),
		jettison.NilMapEmpty(),
	)
}
