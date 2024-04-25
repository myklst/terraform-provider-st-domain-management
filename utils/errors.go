package utils

import (
	"encoding/json"
)

type DefaultResponse struct {
	Message []string    `json:"msg"`
	Error   interface{} `binding:"required" form:"err" json:"err"`
}

// Extracts the error message from the generic HTTP response
func Extract(input []byte) any {
	obj := DefaultResponse{}
	err := json.Unmarshal(input, &obj)

	if err != nil {
		return err
	}

	return obj.Error
}
