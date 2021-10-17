package apiutils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONError struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func SendJSONError(w http.ResponseWriter, code int, msg string, args ...interface{}) {
	b, _ := json.Marshal(JSONError{false, fmt.Sprintf(msg, args...)})
	http.Error(w, string(b), code)
}
