//go:build ignore

package main

import "encoding/json"

// customHandlers registers your Go handlers callable from JavaScript
// via lightshell.invoke(name, payload).
//
// Example:
//
//	Handle("greet", func(payload json.RawMessage) (any, error) {
//	    var p struct { Name string `json:"name"` }
//	    json.Unmarshal(payload, &p)
//	    return map[string]any{"message": "Hello, " + p.Name + "!"}, nil
//	})
//
// From JavaScript:
//
//	const result = await lightshell.invoke("greet", { name: "Alice" })
//	// result = { message: "Hello, Alice!" }
//
// Use OnShutdown to clean up resources when the app exits:
//
//	OnShutdown(func() { fmt.Println("Goodbye!") })
func customHandlers() {
	// Register your custom handlers here.
	_ = json.RawMessage{} // keep import
}
