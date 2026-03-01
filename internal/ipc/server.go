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
	mu              sync.RWMutex
	handlers        map[string]HandlerFunc
	customHandlers  map[string]HandlerFunc
	evalFunc        func(js string) // function to evaluate JS in the webview
	shutdownHooks   []func()
}

// NewRouter creates a new IPC router.
func NewRouter() *Router {
	r := &Router{
		handlers:       make(map[string]HandlerFunc),
		customHandlers: make(map[string]HandlerFunc),
	}
	// Register the invoke dispatcher that routes to custom handlers
	r.handlers["invoke"] = r.handleInvoke
	return r
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

// HandleCustom registers a custom handler invokable from JS via lightshell.invoke(name, payload).
func (r *Router) HandleCustom(name string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.customHandlers[name] = handler
}

// OnShutdown registers a function to be called when the app is shutting down.
func (r *Router) OnShutdown(fn func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shutdownHooks = append(r.shutdownHooks, fn)
}

// RunShutdownHooks calls all registered shutdown hooks.
func (r *Router) RunShutdownHooks() {
	r.mu.RLock()
	hooks := make([]func(), len(r.shutdownHooks))
	copy(hooks, r.shutdownHooks)
	r.mu.RUnlock()
	for _, fn := range hooks {
		fn()
	}
}

// handleInvoke dispatches lightshell.invoke() calls to custom handlers.
func (r *Router) handleInvoke(params json.RawMessage) (any, error) {
	var p struct {
		Handler string          `json:"handler"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid invoke params: %v", err)
	}

	r.mu.RLock()
	handler, ok := r.customHandlers[p.Handler]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown handler: %s", p.Handler)
	}

	return handler(p.Payload)
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
