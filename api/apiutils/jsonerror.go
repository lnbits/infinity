package apiutils

import (
	"fmt"
	"net/http"

	"github.com/lnbits/infinity/utils"
)

type JSONError struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func SendJSONError(w http.ResponseWriter, code int, msg string, args ...interface{}) {
	b, _ := utils.JSONMarshal(JSONError{false, fmt.Sprintf(msg, args...)})
	http.Error(w, string(b), code)
}
