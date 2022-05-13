package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/lnbits/infinity/utils"
)

type StringList []string

func (sl *StringList) Scan(src interface{}) error {
	if jstr, ok := src.(string); ok {
		return json.Unmarshal([]byte(jstr), &sl)
	} else {
		return errors.New("value is not a string")
	}
}

func (sl StringList) Value() (driver.Value, error) {
	if j, err := utils.JSONMarshal(sl); err == nil {
		return string(j), nil
	} else {
		return nil, err
	}
}

type JSONObject map[string]interface{}

func (jo *JSONObject) Scan(src interface{}) error {
	if jstr, ok := src.(string); ok {
		return json.Unmarshal([]byte(jstr), &jo)
	} else {
		return errors.New("value is not a string")
	}
}

func (jo JSONObject) Value() (driver.Value, error) {
	if j, err := utils.JSONMarshal(jo); err == nil {
		return string(j), nil
	} else {
		return nil, err
	}
}
