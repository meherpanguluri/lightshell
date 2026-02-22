package tests

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/lightshell-dev/lightshell/internal/ipc"
)

// parseResponse unmarshals a JSON response string into an ipc.Response.
func parseResponse(t *testing.T, raw string) ipc.Response {
	t.Helper()
	var resp ipc.Response
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
		wantErr  bool
		handler  ipc.HandlerFunc
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
				// Result comes back as string via JSON round-trip
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
				// JSON numbers unmarshal as float64
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
			router := ipc.NewRouter()
			router.Handle(tt.method, tt.handler)

			paramsBytes, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("failed to marshal params: %v", err)
			}

			reqMsg, err := json.Marshal(map[string]any{
				"id":     "test-id-1",
				"method": tt.method,
				"params": json.RawMessage(paramsBytes),
			})
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

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
	router := ipc.NewRouter()

	// Register a handler for a different method
	router.Handle("known.method", func(params json.RawMessage) (any, error) {
		return "ok", nil
	})

	// Call an unknown method
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
		{
			name:   "completely invalid JSON",
			rawMsg: "this is not json",
		},
		{
			name:   "truncated JSON",
			rawMsg: `{"id": "abc", "method":`,
		},
		{
			name:   "empty string",
			rawMsg: "",
		},
		{
			name:   "array instead of object",
			rawMsg: `[1, 2, 3]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := ipc.NewRouter()
			raw := router.HandleMessage(tt.rawMsg)
			resp := parseResponse(t, raw)

			if resp.Error == "" {
				t.Fatal("expected an error for invalid message, got none")
			}
		})
	}
}

func TestIPCConcurrency(t *testing.T) {
	router := ipc.NewRouter()

	// Register a simple handler
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

			var resp ipc.Response
			if err := json.Unmarshal([]byte(raw), &resp); err != nil {
				errors <- "failed to parse response: " + err.Error()
				return
			}
			if resp.Error != "" {
				errors <- "unexpected error: " + resp.Error
				return
			}
			// Result should be a float64 matching our value
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
	router := ipc.NewRouter()

	router.Handle("test.fail", func(params json.RawMessage) (any, error) {
		var dummy int
		return nil, json.Unmarshal([]byte("bad"), &dummy) // produces a real error
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
	router := ipc.NewRouter()

	var capturedJS string
	router.SetEvalFunc(func(js string) {
		capturedJS = js
	})

	router.SendEvent("window.resize", map[string]int{"width": 1024, "height": 768})

	if capturedJS == "" {
		t.Fatal("expected eval function to be called, but it was not")
	}

	// The JS should look like: __lightshell_receive({"event":"window.resize","data":{...}})
	const prefix = "__lightshell_receive("
	const suffix = ")"
	if len(capturedJS) < len(prefix)+len(suffix) {
		t.Fatalf("JS too short: %q", capturedJS)
	}
	if capturedJS[:len(prefix)] != prefix {
		t.Errorf("expected JS to start with %q, got %q", prefix, capturedJS[:len(prefix)])
	}
	if capturedJS[len(capturedJS)-1:] != suffix {
		t.Errorf("expected JS to end with %q, got %q", suffix, capturedJS[len(capturedJS)-1:])
	}

	// Parse the JSON payload inside the function call
	jsonPayload := capturedJS[len(prefix) : len(capturedJS)-len(suffix)]
	var evt ipc.Event
	if err := json.Unmarshal([]byte(jsonPayload), &evt); err != nil {
		t.Fatalf("failed to parse event JSON: %v", err)
	}
	if evt.EventName != "window.resize" {
		t.Errorf("expected event name %q, got %q", "window.resize", evt.EventName)
	}
}

func TestIPCSendEventNoEvalFunc(t *testing.T) {
	// When no eval function is set, SendEvent should not panic
	router := ipc.NewRouter()
	router.SendEvent("test.event", map[string]string{"key": "value"})
	// If we get here without panicking, the test passes
}
