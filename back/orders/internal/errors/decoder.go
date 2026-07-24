package errors

import "encoding/json"

type JsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Error `json:"data,omitempty"`
}

func (err JsonRPCError) Error() string {
	return err.Message
}

func ErrorDecoder(errData json.RawMessage) (err error) {

	var jsonrpcError JsonRPCError
	if err = json.Unmarshal(errData, &jsonrpcError); err != nil {
		return
	}
	return jsonrpcError
}
