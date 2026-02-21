package ipc

import "encoding/json"

// Request is a message from JS to Go.
type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// Response is a message from Go to JS.
type Response struct {
	ID     string `json:"id"`
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
}

// Event is a push message from Go to JS (no request ID).
type Event struct {
	EventName string `json:"event"`
	Data      any    `json:"data"`
}
