package ipc

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func parseResponse(t *testing.T, raw string) Response {
	t.Helper()
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to parse response JSON %q: %v", raw, err)
	}
	return resp
}

func TestIPCRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		params   map[string]any
		handler  HandlerFunc
		checkRes func(t *testing.T, result any)
	}{
		{
			name:   "echo handler returns params",
			method: "test.echo",
			params: map[string]any{"msg": "hello"},
			handler: func(params json.RawMessage) (any, error) {
				var p struct {
					Msg string `json:"msg"`
				}
				json.Unmarshal(params, &p)
				return p.Msg, nil
			},
			checkRes: func(t *testing.T, result any) {
				s, ok := result.(string)
				if !ok {
					t.Fatalf("expected string result, got %T: %v", result, result)
				}
				if s != "hello" {
					t.Errorf("expected %q, got %q", "hello", s)
				}
			},
		},
		{
			name:   "numeric handler returns number",
			method: "test.add",
			params: map[string]any{"a": 2, "b": 3},
			handler: func(params json.RawMessage) (any, error) {
				var p struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				json.Unmarshal(params, &p)
				return p.A + p.B, nil
			},
			checkRes: func(t *testing.T, result any) {
				n, ok := result.(float64)
				if !ok {
					t.Fatalf("expected float64 result, got %T: %v", result, result)
				}
				if n != 5 {
					t.Errorf("expected 5, got %v", n)
				}
			},
		},
		{
			name:   "handler returns nil result",
			method: "test.noop",
			params: map[string]any{},
			handler: func(params json.RawMessage) (any, error) {
				return nil, nil
			},
			checkRes: func(t *testing.T, result any) {
				if result != nil {
					t.Errorf("expected nil result, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			router.Handle(tt.method, tt.handler)

			paramsBytes, _ := json.Marshal(tt.params)
			reqMsg, _ := json.Marshal(map[string]any{
				"id":     "test-id-1",
				"method": tt.method,
				"params": json.RawMessage(paramsBytes),
			})

			raw := router.HandleMessage(string(reqMsg))
			resp := parseResponse(t, raw)

			if resp.ID != "test-id-1" {
				t.Errorf("expected response ID %q, got %q", "test-id-1", resp.ID)
			}
			if resp.Error != "" {
				t.Errorf("expected no error, got %q", resp.Error)
			}
			if tt.checkRes != nil {
				tt.checkRes(t, resp.Result)
			}
		})
	}
}

func TestIPCUnknownMethod(t *testing.T) {
	router := NewRouter()
	router.Handle("known.method", func(params json.RawMessage) (any, error) {
		return "ok", nil
	})

	reqMsg := `{"id":"req-unknown","method":"unknown.method","params":{}}`
	raw := router.HandleMessage(reqMsg)
	resp := parseResponse(t, raw)

	if resp.ID != "req-unknown" {
		t.Errorf("expected response ID %q, got %q", "req-unknown", resp.ID)
	}
	if resp.Error == "" {
		t.Fatal("expected an error for unknown method, got none")
	}
	if resp.Result != nil {
		t.Errorf("expected nil result for error response, got %v", resp.Result)
	}
}

func TestIPCInvalidMessage(t *testing.T) {
	tests := []struct {
		name   string
		rawMsg string
	}{
		{"completely invalid JSON", "this is not json"},
		{"truncated JSON", `{"id": "abc", "method":`},
		{"empty string", ""},
		{"array instead of object", `[1, 2, 3]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			raw := router.HandleMessage(tt.rawMsg)
			resp := parseResponse(t, raw)
			if resp.Error == "" {
				t.Fatal("expected an error for invalid message, got none")
			}
		})
	}
}

func TestIPCConcurrency(t *testing.T) {
	router := NewRouter()
	router.Handle("concurrent.echo", func(params json.RawMessage) (any, error) {
		var p struct {
			Value int `json:"value"`
		}
		json.Unmarshal(params, &p)
		return p.Value, nil
	})

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(val int) {
			defer wg.Done()
			reqID := fmt.Sprintf("concurrent-%d", val)
			paramsBytes, _ := json.Marshal(map[string]any{"value": val})
			reqMsg, _ := json.Marshal(map[string]any{
				"id":     reqID,
				"method": "concurrent.echo",
				"params": json.RawMessage(paramsBytes),
			})
			raw := router.HandleMessage(string(reqMsg))
			var resp Response
			if err := json.Unmarshal([]byte(raw), &resp); err != nil {
				errors <- "failed to parse response: " + err.Error()
				return
			}
			if resp.Error != "" {
				errors <- "unexpected error: " + resp.Error
				return
			}
			result, ok := resp.Result.(float64)
			if !ok {
				errors <- "result is not float64"
				return
			}
			if int(result) != val {
				errors <- "result mismatch"
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	for errMsg := range errors {
		t.Errorf("concurrent error: %s", errMsg)
	}
}

func TestIPCHandlerError(t *testing.T) {
	router := NewRouter()
	router.Handle("test.fail", func(params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("intentional failure")
	})

	reqMsg := `{"id":"req-fail","method":"test.fail","params":{}}`
	raw := router.HandleMessage(reqMsg)
	resp := parseResponse(t, raw)

	if resp.ID != "req-fail" {
		t.Errorf("expected response ID %q, got %q", "req-fail", resp.ID)
	}
	if resp.Error == "" {
		t.Fatal("expected an error from failing handler, got none")
	}
}

func TestIPCSendEvent(t *testing.T) {
	router := NewRouter()
	var capturedJS string
	router.SetEvalFunc(func(js string) {
		capturedJS = js
	})

	router.SendEvent("window.resize", map[string]int{"width": 1024, "height": 768})

	if capturedJS == "" {
		t.Fatal("expected eval function to be called, but it was not")
	}

	const prefix = "__lightshell_receive("
	const suffix = ")"
	if len(capturedJS) < len(prefix)+len(suffix) {
		t.Fatalf("JS too short: %q", capturedJS)
	}
	if capturedJS[:len(prefix)] != prefix {
		t.Errorf("expected JS to start with %q", prefix)
	}

	jsonPayload := capturedJS[len(prefix) : len(capturedJS)-len(suffix)]
	var evt Event
	if err := json.Unmarshal([]byte(jsonPayload), &evt); err != nil {
		t.Fatalf("failed to parse event JSON: %v", err)
	}
	if evt.EventName != "window.resize" {
		t.Errorf("expected event name %q, got %q", "window.resize", evt.EventName)
	}
}

func TestIPCSendEventNoEvalFunc(t *testing.T) {
	router := NewRouter()
	// Should not panic
	router.SendEvent("test.event", map[string]string{"key": "value"})
}

// --- Tests for new invoke/custom handler/shutdown functionality ---

func TestHandleCustomAndInvoke(t *testing.T) {
	router := NewRouter()

	router.HandleCustom("ai.status", func(params json.RawMessage) (any, error) {
		return map[string]any{"ready": true, "model": "test-v1"}, nil
	})

	// Call invoke with the handler name
	paramsBytes, _ := json.Marshal(map[string]any{
		"handler": "ai.status",
		"payload": map[string]any{},
	})
	reqMsg, _ := json.Marshal(map[string]any{
		"id":     "invoke-1",
		"method": "invoke",
		"params": json.RawMessage(paramsBytes),
	})

	raw := router.HandleMessage(string(reqMsg))
	resp := parseResponse(t, raw)

	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if result["ready"] != true {
		t.Errorf("expected ready=true, got %v", result["ready"])
	}
	if result["model"] != "test-v1" {
		t.Errorf("expected model=test-v1, got %v", result["model"])
	}
}

func TestInvokeWithPayload(t *testing.T) {
	router := NewRouter()

	router.HandleCustom("greet", func(params json.RawMessage) (any, error) {
		var p struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return map[string]string{"message": "Hello, " + p.Name + "!"}, nil
	})

	paramsBytes, _ := json.Marshal(map[string]any{
		"handler": "greet",
		"payload": map[string]any{"name": "Alice"},
	})
	reqMsg, _ := json.Marshal(map[string]any{
		"id":     "invoke-2",
		"method": "invoke",
		"params": json.RawMessage(paramsBytes),
	})

	raw := router.HandleMessage(string(reqMsg))
	resp := parseResponse(t, raw)

	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	if result["message"] != "Hello, Alice!" {
		t.Errorf("expected 'Hello, Alice!', got %v", result["message"])
	}
}

func TestInvokeUnknownHandler(t *testing.T) {
	router := NewRouter()

	paramsBytes, _ := json.Marshal(map[string]any{
		"handler": "nonexistent",
		"payload": map[string]any{},
	})
	reqMsg, _ := json.Marshal(map[string]any{
		"id":     "invoke-3",
		"method": "invoke",
		"params": json.RawMessage(paramsBytes),
	})

	raw := router.HandleMessage(string(reqMsg))
	resp := parseResponse(t, raw)

	if resp.Error == "" {
		t.Fatal("expected error for unknown handler, got none")
	}
	if resp.ID != "invoke-3" {
		t.Errorf("expected response ID %q, got %q", "invoke-3", resp.ID)
	}
}

func TestInvokeInvalidParams(t *testing.T) {
	router := NewRouter()

	reqMsg, _ := json.Marshal(map[string]any{
		"id":     "invoke-4",
		"method": "invoke",
		"params": json.RawMessage(`"not an object"`),
	})

	raw := router.HandleMessage(string(reqMsg))
	resp := parseResponse(t, raw)

	if resp.Error == "" {
		t.Fatal("expected error for invalid invoke params, got none")
	}
}

func TestInvokeHandlerError(t *testing.T) {
	router := NewRouter()

	router.HandleCustom("fail", func(params json.RawMessage) (any, error) {
		return nil, fmt.Errorf("handler crashed")
	})

	paramsBytes, _ := json.Marshal(map[string]any{
		"handler": "fail",
		"payload": map[string]any{},
	})
	reqMsg, _ := json.Marshal(map[string]any{
		"id":     "invoke-5",
		"method": "invoke",
		"params": json.RawMessage(paramsBytes),
	})

	raw := router.HandleMessage(string(reqMsg))
	resp := parseResponse(t, raw)

	if resp.Error == "" {
		t.Fatal("expected error from failing handler")
	}
}

func TestOnShutdownAndRunShutdownHooks(t *testing.T) {
	router := NewRouter()

	var called1, called2 int32
	router.OnShutdown(func() { atomic.AddInt32(&called1, 1) })
	router.OnShutdown(func() { atomic.AddInt32(&called2, 1) })

	router.RunShutdownHooks()

	if atomic.LoadInt32(&called1) != 1 {
		t.Errorf("expected first shutdown hook to be called once, got %d", called1)
	}
	if atomic.LoadInt32(&called2) != 1 {
		t.Errorf("expected second shutdown hook to be called once, got %d", called2)
	}
}

func TestRunShutdownHooksEmpty(t *testing.T) {
	router := NewRouter()
	// Should not panic with no hooks registered
	router.RunShutdownHooks()
}

func TestRunShutdownHooksMultipleTimes(t *testing.T) {
	router := NewRouter()

	var count int32
	router.OnShutdown(func() { atomic.AddInt32(&count, 1) })

	router.RunShutdownHooks()
	router.RunShutdownHooks()

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected shutdown hook to be called twice, got %d", count)
	}
}

func TestSetupRouter(t *testing.T) {
	r := SetupRouter()
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	if r.handlers == nil {
		t.Fatal("expected handlers map to be initialized")
	}
	if r.customHandlers == nil {
		t.Fatal("expected customHandlers map to be initialized")
	}
}

func TestNewRouterHasInvokeHandler(t *testing.T) {
	r := NewRouter()
	if _, ok := r.handlers["invoke"]; !ok {
		t.Fatal("expected 'invoke' handler to be registered by default")
	}
}

func TestHandleCustomConcurrency(t *testing.T) {
	router := NewRouter()

	// Register handlers concurrently
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("handler.%d", n)
			router.HandleCustom(name, func(params json.RawMessage) (any, error) {
				return n, nil
			})
		}(i)
	}
	wg.Wait()

	// Invoke each handler
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("handler.%d", i)
		paramsBytes, _ := json.Marshal(map[string]any{
			"handler": name,
			"payload": map[string]any{},
		})
		reqMsg, _ := json.Marshal(map[string]any{
			"id":     fmt.Sprintf("invoke-%d", i),
			"method": "invoke",
			"params": json.RawMessage(paramsBytes),
		})

		raw := router.HandleMessage(string(reqMsg))
		resp := parseResponse(t, raw)
		if resp.Error != "" {
			t.Errorf("unexpected error for handler.%d: %s", i, resp.Error)
		}
	}
}
