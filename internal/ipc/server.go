package ipc

import (
	"encoding/json"
	"fmt"
	"sync"
)

// HandlerFunc processes an IPC request and returns a result or error.
type HandlerFunc func(params json.RawMessage) (any, error)

// Router routes IPC method calls to handler functions.
type Router struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
	evalFunc func(js string) // function to evaluate JS in the webview
}

// NewRouter creates a new IPC router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]HandlerFunc),
	}
}

// SetEvalFunc sets the function used to send JS to the webview.
func (r *Router) SetEvalFunc(fn func(js string)) {
	r.evalFunc = fn
}

// Handle registers a handler for a method name.
func (r *Router) Handle(method string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = handler
}

// HandleMessage processes a raw JSON message from the webview and returns the JSON response.
func (r *Router) HandleMessage(rawMsg string) string {
	var req Request
	if err := json.Unmarshal([]byte(rawMsg), &req); err != nil {
		return errorResponse("", fmt.Sprintf("invalid message: %v", err))
	}

	r.mu.RLock()
	handler, ok := r.handlers[req.Method]
	r.mu.RUnlock()

	if !ok {
		return errorResponse(req.ID, fmt.Sprintf("unknown method: %s", req.Method))
	}

	result, err := handler(req.Params)
	if err != nil {
		return errorResponse(req.ID, err.Error())
	}

	return successResponse(req.ID, result)
}

// SendEvent sends an event to the webview via JS eval.
func (r *Router) SendEvent(eventName string, data any) {
	if r.evalFunc == nil {
		return
	}
	evt := Event{EventName: eventName, Data: data}
	jsonBytes, err := json.Marshal(evt)
	if err != nil {
		return
	}
	js := fmt.Sprintf("__lightshell_receive(%s)", string(jsonBytes))
	r.evalFunc(js)
}

func successResponse(id string, result any) string {
	resp := Response{ID: id, Result: result}
	data, _ := json.Marshal(resp)
	return string(data)
}

func errorResponse(id string, errMsg string) string {
	resp := Response{ID: id, Error: errMsg}
	data, _ := json.Marshal(resp)
	return string(data)
}
