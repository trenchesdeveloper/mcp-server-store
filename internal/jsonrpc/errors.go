package jsonrpc

// JSON-RPC 2.0 Error Codes
const (
	ErrorParse     = -32700 // Invalid JSON
	ErrorInvalidRequest = -32600 // Invalid Request
	ErrorMethodNotFound = -32601 // Method Not Found
	ErrorInvalidParams  = -32602 // Invalid Params
	ErrorInternal  = -32603 // Internal Error
)