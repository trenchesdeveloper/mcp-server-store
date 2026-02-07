package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 Request
type Request struct {
	JSONRPC string      `json:"jsonrpc"` // Version of the JSON-RPC protocol
	Method  string      `json:"method"` // Method to be invoked
	Params  json.RawMessage `json:"params,omitempty"` // Parameters to be passed to the method
	ID      interface{} `json:"id,omitempty"` // Identifier of the request
}

// JSON-RPC 2.0 Response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// JSON-RPC 2.0 Error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}


func (e *Error) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s, Data: %v", e.Code, e.Message, e.Data)
}

func (r *Request) IsNotification() bool {
	return r.ID == nil
}

func (r *Request) Validate() error {
	if r.JSONRPC != "2.0" {
		return &Error{
			Code: ErrorInvalidRequest,
			Message: "Invalid JSON-RPC version",
		}
	}
	if r.Method == "" {
		return &Error{
			Code: ErrorInvalidRequest,
			Message: "Method is required",
		}
	}
	return nil
}

// NewResponse creates a new JSON-RPC response
func NewResponse(id interface{}, result interface{}, err *Error) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   err,
	}
}

// NewErrorResponse creates a new JSON-RPC error response
func NewErrorResponse(id interface{}, err *Error) *Response {
	return NewResponse(id, nil, err)
}

// NewSuccessResponse creates a new JSON-RPC success response
func NewSuccessResponse(id interface{}, result interface{}) *Response {
	return NewResponse(id, result, nil)
}