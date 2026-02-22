package cli

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/lightshell-dev/lightshell/internal/webview"
)

// mcpSocketServer handles commands from the MCP server process via a Unix
// domain socket. It runs inside the lightshell dev process.
type mcpSocketServer struct {
	socketPath  string
	listener    net.Listener
	wv          webview.Webview
	console     *mcpConsoleBuffer
	mu          sync.Mutex
	evalResults map[string]chan evalResult
	closed      bool
}

// evalResult holds the result (or error) from a JS evaluation.
type evalResult struct {
	Value string
	Error string
}

// mcpConsoleBuffer is a thread-safe ring buffer for console log entries
// within the dev process.
type mcpConsoleBuffer struct {
	mu      sync.Mutex
	entries []mcpConsoleEntry
	max     int
}

// mcpConsoleEntry represents a single console log entry.
type mcpConsoleEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// mcpSocketCommand is the JSON command received from the MCP server.
type mcpSocketCommand struct {
	ID       int    `json:"id"`
	Cmd      string `json:"cmd"`
	Delay    int    `json:"delay,omitempty"`
	Lines    int    `json:"lines,omitempty"`
	Level    string `json:"level,omitempty"`
	Clear    bool   `json:"clear,omitempty"`
	Selector string `json:"selector,omitempty"`
	Depth    int    `json:"depth,omitempty"`
	Code     string `json:"code,omitempty"`
}

// mcpSocketResponse is the JSON response sent back to the MCP server.
type mcpSocketResponse struct {
	ID      int               `json:"id"`
	Error   string            `json:"error,omitempty"`
	Result  json.RawMessage   `json:"result,omitempty"`
	Image   string            `json:"image,omitempty"`
	Width   int               `json:"width,omitempty"`
	Height  int               `json:"height,omitempty"`
	Status  string            `json:"status,omitempty"`
	HTML    string            `json:"html,omitempty"`
	Entries []mcpConsoleEntry `json:"entries,omitempty"`
}

// newMCPSocketServer creates a new MCP socket server.
func newMCPSocketServer(socketPath string, wv webview.Webview) *mcpSocketServer {
	return &mcpSocketServer{
		socketPath: socketPath,
		wv:         wv,
		console: &mcpConsoleBuffer{
			entries: make([]mcpConsoleEntry, 0, 1000),
			max:     1000,
		},
		evalResults: make(map[string]chan evalResult),
	}
}

// serve starts the Unix domain socket server. It accepts one connection at a
// time (the MCP server is the only client) and processes commands sequentially.
func (s *mcpSocketServer) serve() error {
	// Remove any stale socket file
	os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.socketPath, err)
	}
	s.listener = listener

	// Set socket permissions to owner-only (0600)
	os.Chmod(s.socketPath, 0600)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil
			}
			fmt.Fprintf(os.Stderr, "[mcp-socket] accept error: %v\n", err)
			continue
		}

		s.handleConnection(conn)
	}
}

// handleConnection processes commands from a single MCP server connection.
func (s *mcpSocketServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	// Allow up to 10MB per line for large responses (screenshots)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var cmd mcpSocketCommand
		if err := json.Unmarshal(line, &cmd); err != nil {
			resp := mcpSocketResponse{Error: fmt.Sprintf("invalid command: %v", err)}
			s.writeResponse(conn, resp)
			continue
		}

		resp := s.handleCommand(cmd)
		s.writeResponse(conn, resp)
	}
}

// writeResponse writes a JSON response followed by a newline to the connection.
func (s *mcpSocketServer) writeResponse(conn net.Conn, resp mcpSocketResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		// Last resort: write a plain error
		data = []byte(`{"error":"failed to marshal response"}`)
	}
	data = append(data, '\n')
	conn.Write(data)
}

// handleCommand dispatches a command and returns the response.
func (s *mcpSocketServer) handleCommand(cmd mcpSocketCommand) mcpSocketResponse {
	switch cmd.Cmd {
	case "screenshot":
		return s.handleScreenshot(cmd)
	case "console":
		return s.handleConsole(cmd)
	case "eval":
		return s.handleEval(cmd)
	case "dom":
		return s.handleDOM(cmd)
	case "reload":
		return s.handleReload(cmd)
	default:
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: fmt.Sprintf("unknown command: %s", cmd.Cmd),
		}
	}
}

// handleScreenshot captures a screenshot of the webview.
func (s *mcpSocketServer) handleScreenshot(cmd mcpSocketCommand) mcpSocketResponse {
	// Wait for the specified delay (allows animations/rendering to complete)
	delay := cmd.Delay
	if delay > 5000 {
		delay = 5000 // cap at 5 seconds
	}
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	data, err := s.wv.Screenshot()
	if err != nil {
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: fmt.Sprintf("screenshot failed: %v", err),
		}
	}

	// Cap screenshot size at 5MB to prevent oversized responses
	const maxScreenshotSize = 5 * 1024 * 1024
	if len(data) > maxScreenshotSize {
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: fmt.Sprintf("screenshot too large: %d bytes (max %d)", len(data), maxScreenshotSize),
		}
	}

	// Get window dimensions for the response
	width, height := s.wv.GetSize()

	encoded := base64.StdEncoding.EncodeToString(data)
	return mcpSocketResponse{
		ID:     cmd.ID,
		Image:  encoded,
		Width:  width,
		Height: height,
	}
}

// handleConsole returns console log entries from the buffer.
func (s *mcpSocketServer) handleConsole(cmd mcpSocketCommand) mcpSocketResponse {
	lines := cmd.Lines
	if lines <= 0 {
		lines = 50 // default
	}
	level := cmd.Level
	if level == "" {
		level = "all"
	}

	entries := s.console.get(lines, level)

	if cmd.Clear {
		s.console.clear()
	}

	return mcpSocketResponse{
		ID:      cmd.ID,
		Entries: entries,
	}
}

// handleEval evaluates JavaScript code in the webview and returns the result.
func (s *mcpSocketServer) handleEval(cmd mcpSocketCommand) mcpSocketResponse {
	if cmd.Code == "" {
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: "eval: code is required",
		}
	}

	// Generate a unique callback ID
	callbackID := fmt.Sprintf("mcp_eval_%d_%d", cmd.ID, time.Now().UnixNano())

	// Create a channel for the result
	resultCh := make(chan evalResult, 1)
	s.mu.Lock()
	s.evalResults[callbackID] = resultCh
	s.mu.Unlock()

	// Clean up the channel when done
	defer func() {
		s.mu.Lock()
		delete(s.evalResults, callbackID)
		s.mu.Unlock()
	}()

	// Build JS that evaluates the code and sends the result back via postMessage.
	// We JSON-encode both the user code and the callback ID to prevent injection.
	codeJSON, _ := json.Marshal(cmd.Code)
	callbackJSON, _ := json.Marshal(callbackID)
	js := fmt.Sprintf(`(function(){
		try {
			var __r = eval(%s);
			var __v;
			if (__r === undefined) { __v = "undefined"; }
			else if (__r === null) { __v = "null"; }
			else { try { __v = JSON.stringify(__r); } catch(e) { __v = String(__r); } }
			window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
				__mcp_eval: %s,
				result: __v
			}));
		} catch(e) {
			window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
				__mcp_eval: %s,
				error: e.message || String(e)
			}));
		}
	})()`, string(codeJSON), string(callbackJSON), string(callbackJSON))

	// Evaluate the JS in the webview
	if err := s.wv.Eval(js); err != nil {
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: fmt.Sprintf("eval failed: %v", err),
		}
	}

	// Wait for the result with a timeout
	select {
	case result := <-resultCh:
		if result.Error != "" {
			return mcpSocketResponse{
				ID:    cmd.ID,
				Error: fmt.Sprintf("JS error: %s", result.Error),
			}
		}
		// result.Value is already a string from JS (e.g., JSON.stringify output
		// for objects, or a plain string like "undefined"/"null").
		// Wrap it as a proper JSON string value.
		valueJSON, _ := json.Marshal(result.Value)
		return mcpSocketResponse{
			ID:     cmd.ID,
			Result: json.RawMessage(valueJSON),
		}
	case <-time.After(5 * time.Second):
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: "eval timed out after 5s",
		}
	}
}

// handleDOM serializes a portion of the DOM and returns it as HTML.
func (s *mcpSocketServer) handleDOM(cmd mcpSocketCommand) mcpSocketResponse {
	selector := cmd.Selector
	if selector == "" {
		selector = "body"
	}

	depth := cmd.Depth
	if depth <= 0 {
		depth = 3 // default depth
	}

	// Generate a unique callback ID for this DOM request
	callbackID := fmt.Sprintf("mcp_dom_%d_%d", cmd.ID, time.Now().UnixNano())

	// Build JS to serialize the DOM at the given selector and depth
	selectorJSON, _ := json.Marshal(selector)
	callbackJSON, _ := json.Marshal(callbackID)
	js := fmt.Sprintf(`(function(){
		function serializeDOM(el, depth) {
			if (!el || depth <= 0) return '';
			var tag = el.tagName ? el.tagName.toLowerCase() : '';
			if (!tag) return el.textContent || '';
			var attrs = '';
			if (el.attributes) {
				for (var i = 0; i < el.attributes.length; i++) {
					var a = el.attributes[i];
					attrs += ' ' + a.name + '="' + a.value.replace(/"/g, '&quot;') + '"';
				}
			}
			var children = '';
			if (depth > 1 && el.childNodes) {
				for (var j = 0; j < el.childNodes.length; j++) {
					var child = el.childNodes[j];
					if (child.nodeType === 1) {
						children += serializeDOM(child, depth - 1);
					} else if (child.nodeType === 3) {
						var text = child.textContent.trim();
						if (text) children += text;
					}
				}
			} else if (el.childNodes && el.childNodes.length > 0) {
				children = '...';
			}
			var selfClosing = ['br','hr','img','input','meta','link','area','base','col','embed','source','track','wbr'];
			if (selfClosing.indexOf(tag) >= 0) return '<' + tag + attrs + '/>';
			return '<' + tag + attrs + '>' + children + '</' + tag + '>';
		}
		var el = document.querySelector(%s);
		if (!el) {
			window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
				__mcp_eval: %s,
				error: "Element not found: " + %s
			}));
			return;
		}
		var html = serializeDOM(el, %d);
		window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
			__mcp_eval: %s,
			result: html
		}));
	})()`, string(selectorJSON), string(callbackJSON), string(selectorJSON), depth, string(callbackJSON))

	// Create a channel for the DOM result
	resultCh := make(chan evalResult, 1)
	s.mu.Lock()
	s.evalResults[callbackID] = resultCh
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.evalResults, callbackID)
		s.mu.Unlock()
	}()

	if err := s.wv.Eval(js); err != nil {
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: fmt.Sprintf("DOM inspection failed: %v", err),
		}
	}

	select {
	case result := <-resultCh:
		if result.Error != "" {
			return mcpSocketResponse{
				ID:    cmd.ID,
				Error: result.Error,
			}
		}
		return mcpSocketResponse{
			ID:   cmd.ID,
			HTML: result.Value,
		}
	case <-time.After(5 * time.Second):
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: "DOM inspection timed out after 5s",
		}
	}
}

// handleReload triggers a page reload in the webview.
func (s *mcpSocketServer) handleReload(cmd mcpSocketCommand) mcpSocketResponse {
	if err := s.wv.Eval("location.reload()"); err != nil {
		return mcpSocketResponse{
			ID:    cmd.ID,
			Error: fmt.Sprintf("reload failed: %v", err),
		}
	}
	return mcpSocketResponse{
		ID:     cmd.ID,
		Status: "ok",
	}
}

// handleMCPMessage checks if a message from the webview is an MCP-specific
// message (console forwarding or eval result). Returns true if the message
// was handled and should not be routed to the normal IPC handler.
func (s *mcpSocketServer) handleMCPMessage(msg string) bool {
	// Try to parse as JSON to check for MCP-specific fields
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(msg), &obj); err != nil {
		return false
	}

	// Check for console forwarding messages
	if _, ok := obj["__mcp_console"]; ok {
		var entry struct {
			Level   string `json:"level"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(msg), &entry); err != nil {
			return true // still an MCP message, just malformed
		}
		msg := entry.Message
		// Cap console message size to 10KB to prevent memory abuse
		const maxMsgSize = 10 * 1024
		if len(msg) > maxMsgSize {
			msg = msg[:maxMsgSize] + "... (truncated)"
		}
		s.console.add(mcpConsoleEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     entry.Level,
			Message:   msg,
		})
		return true
	}

	// Check for eval/DOM result messages
	if evalIDRaw, ok := obj["__mcp_eval"]; ok {
		var evalID string
		if err := json.Unmarshal(evalIDRaw, &evalID); err != nil {
			return true
		}

		s.mu.Lock()
		ch, found := s.evalResults[evalID]
		s.mu.Unlock()

		if found {
			var resp struct {
				Result string `json:"result"`
				Error  string `json:"error"`
			}
			json.Unmarshal([]byte(msg), &resp)

			// Non-blocking send (channel is buffered with capacity 1)
			select {
			case ch <- evalResult{Value: resp.Result, Error: resp.Error}:
			default:
			}
		}
		return true
	}

	return false
}

// close shuts down the MCP socket server and cleans up resources.
func (s *mcpSocketServer) close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.socketPath)
}

// mcpConsoleBuffer methods

func (b *mcpConsoleBuffer) add(entry mcpConsoleEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) >= b.max {
		copy(b.entries, b.entries[1:])
		b.entries = b.entries[:len(b.entries)-1]
	}
	b.entries = append(b.entries, entry)
}

func (b *mcpConsoleBuffer) get(n int, level string) []mcpConsoleEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	if level == "" || level == "all" {
		if n <= 0 || n >= len(b.entries) {
			result := make([]mcpConsoleEntry, len(b.entries))
			copy(result, b.entries)
			return result
		}
		start := len(b.entries) - n
		result := make([]mcpConsoleEntry, n)
		copy(result, b.entries[start:])
		return result
	}

	var filtered []mcpConsoleEntry
	for _, e := range b.entries {
		if e.Level == level {
			filtered = append(filtered, e)
		}
	}

	if n <= 0 || n >= len(filtered) {
		return filtered
	}

	start := len(filtered) - n
	return filtered[start:]
}

func (b *mcpConsoleBuffer) clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = b.entries[:0]
}

// mcpConsoleForwardScript is injected into the webview when running in MCP mode.
// It wraps console.log/warn/error/info/debug to forward entries to Go via
// postMessage, and captures unhandled errors and promise rejections.
const mcpConsoleForwardScript = `(function(){
	var orig = {
		log: console.log,
		warn: console.warn,
		error: console.error,
		info: console.info,
		debug: console.debug
	};
	['log','warn','error','info','debug'].forEach(function(level){
		console[level] = function(){
			orig[level].apply(console, arguments);
			var args = Array.prototype.slice.call(arguments).map(function(a){
				if (a === null) return 'null';
				if (a === undefined) return 'undefined';
				if (typeof a === 'object') {
					try { return JSON.stringify(a); } catch(e) { return String(a); }
				}
				return String(a);
			});
			try {
				window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
					__mcp_console: true,
					level: level,
					message: args.join(' ')
				}));
			} catch(e) {}
		};
	});
	window.addEventListener('error', function(e) {
		try {
			window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
				__mcp_console: true,
				level: 'error',
				message: e.message + (e.filename ? ' at ' + e.filename + ':' + e.lineno : '')
			}));
		} catch(ex) {}
	});
	window.addEventListener('unhandledrejection', function(e) {
		try {
			window.webkit.messageHandlers.lightshell.postMessage(JSON.stringify({
				__mcp_console: true,
				level: 'error',
				message: 'Unhandled rejection: ' + (e.reason instanceof Error ? e.reason.message : String(e.reason))
			}));
		} catch(ex) {}
	});
})();`
