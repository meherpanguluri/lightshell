package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"sync"
)

// Tool defines an MCP tool with its schema and handler function.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	Handler     func(params map[string]any) (any, error) `json:"-"`
}

// Resource defines an MCP resource with its URI and handler function.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
	Handler     func() (string, error) `json:"-"`
}

// DevProcess is an alias kept for backward compatibility within this package.
// The real implementation is DevProcessManager in devprocess.go.

// JSON-RPC 2.0 types

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server is the MCP JSON-RPC 2.0 server that communicates over stdio.
type Server struct {
	projectDir string
	apiDocs    string
	devProcess *DevProcessManager
	tools      map[string]Tool
	resources  map[string]Resource
	logger     *log.Logger
	writer     io.Writer
	mu         sync.Mutex
}

// NewServer creates a new MCP server for the given project directory.
// apiDocs is the full LightShell API reference text to expose as a resource.
func NewServer(projectDir string, apiDocs string) *Server {
	s := &Server{
		projectDir: projectDir,
		apiDocs:    apiDocs,
		devProcess: NewDevProcessManager(projectDir),
		tools:      make(map[string]Tool),
		resources:  make(map[string]Resource),
		logger:     log.New(os.Stderr, "[lightshell-mcp] ", log.LstdFlags),
		writer:     os.Stdout,
	}
	s.registerTools()
	s.registerResources()
	return s
}

// Run starts the MCP server main loop, reading JSON-RPC requests from stdin
// and writing responses to stdout.
func (s *Server) Run() error {
	s.logger.Println("MCP server starting")

	scanner := bufio.NewScanner(os.Stdin)
	// Allow up to 10MB per line for large messages
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.logger.Printf("Failed to parse request: %v", err)
			// If we can't parse the request, we can't know the ID.
			// Send a parse error with null ID.
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		s.logger.Printf("Received: method=%s id=%v", req.Method, req.ID)
		s.handleRequest(req)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdin scanner error: %w", err)
	}

	s.logger.Println("MCP server shutting down (stdin closed)")
	return nil
}

func (s *Server) handleRequest(req jsonRPCRequest) {
	// Notifications have no ID and expect no response
	if req.ID == nil {
		s.handleNotification(req)
		return
	}

	var result any
	var rpcErr *jsonRPCError

	switch req.Method {
	case "initialize":
		result = s.handleInitialize(req.Params)
	case "ping":
		result = map[string]any{}
	case "tools/list":
		result = s.handleToolsList()
	case "tools/call":
		result, rpcErr = s.handleToolsCall(req.Params)
	case "resources/list":
		result = s.handleResourcesList()
	case "resources/read":
		result, rpcErr = s.handleResourcesRead(req.Params)
	default:
		rpcErr = &jsonRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	if rpcErr != nil {
		s.sendError(req.ID, rpcErr.Code, rpcErr.Message)
	} else {
		s.sendResult(req.ID, result)
	}
}

func (s *Server) handleNotification(req jsonRPCRequest) {
	switch req.Method {
	case "notifications/initialized":
		s.logger.Println("Client initialized")
	default:
		s.logger.Printf("Unknown notification: %s", req.Method)
	}
}

func (s *Server) handleInitialize(params json.RawMessage) any {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "lightshell",
			"version": "0.1.0",
		},
	}
}

func (s *Server) handleToolsList() any {
	// Collect and sort tool names for deterministic ordering
	names := make([]string, 0, len(s.tools))
	for name := range s.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	tools := make([]map[string]any, 0, len(s.tools))
	for _, name := range names {
		t := s.tools[name]
		tools = append(tools, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return map[string]any{
		"tools": tools,
	}
}

func (s *Server) handleToolsCall(params json.RawMessage) (any, *jsonRPCError) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &jsonRPCError{
			Code:    -32602,
			Message: fmt.Sprintf("Invalid params: %v", err),
		}
	}

	tool, ok := s.tools[p.Name]
	if !ok {
		return nil, &jsonRPCError{
			Code:    -32602,
			Message: fmt.Sprintf("Unknown tool: %s", p.Name),
		}
	}

	s.logger.Printf("Calling tool: %s", p.Name)

	result, err := tool.Handler(p.Arguments)
	if err != nil {
		// Tool errors are returned as successful responses with isError flag
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": fmt.Sprintf("Error: %v", err),
				},
			},
			"isError": true,
		}, nil
	}

	// Check if the result is already formatted (e.g., for screenshots with image content)
	if m, ok := result.(map[string]any); ok {
		if _, hasContent := m["content"]; hasContent {
			return m, nil
		}
	}

	// Default: wrap result as text content
	var text string
	switch v := result.(type) {
	case string:
		text = v
	default:
		b, _ := json.Marshal(v)
		text = string(b)
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
	}, nil
}

func (s *Server) handleResourcesList() any {
	// Collect and sort resource URIs for deterministic ordering
	uris := make([]string, 0, len(s.resources))
	for uri := range s.resources {
		uris = append(uris, uri)
	}
	sort.Strings(uris)

	resources := make([]map[string]any, 0, len(s.resources))
	for _, uri := range uris {
		r := s.resources[uri]
		resources = append(resources, map[string]any{
			"uri":         r.URI,
			"name":        r.Name,
			"description": r.Description,
			"mimeType":    r.MimeType,
		})
	}
	return map[string]any{
		"resources": resources,
	}
}

func (s *Server) handleResourcesRead(params json.RawMessage) (any, *jsonRPCError) {
	var p struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &jsonRPCError{
			Code:    -32602,
			Message: fmt.Sprintf("Invalid params: %v", err),
		}
	}

	resource, ok := s.resources[p.URI]
	if !ok {
		return nil, &jsonRPCError{
			Code:    -32602,
			Message: fmt.Sprintf("Unknown resource: %s", p.URI),
		}
	}

	content, err := resource.Handler()
	if err != nil {
		return nil, &jsonRPCError{
			Code:    -32603,
			Message: fmt.Sprintf("Failed to read resource: %v", err),
		}
	}

	return map[string]any{
		"contents": []map[string]any{
			{
				"uri":      resource.URI,
				"mimeType": resource.MimeType,
				"text":     content,
			},
		},
	}, nil
}

// registerTool adds a tool to the server.
func (s *Server) registerTool(t Tool) {
	s.tools[t.Name] = t
}

// registerResource adds a resource to the server.
func (s *Server) registerResource(r Resource) {
	s.resources[r.URI] = r
}

// sendResult writes a successful JSON-RPC response to stdout.
func (s *Server) sendResult(id any, result any) {
	s.send(jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

// sendError writes an error JSON-RPC response to stdout.
func (s *Server) sendError(id any, code int, message string) {
	s.send(jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonRPCError{
			Code:    code,
			Message: message,
		},
	})
}

func (s *Server) send(resp jsonRPCResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		s.logger.Printf("Failed to marshal response: %v", err)
		return
	}

	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		s.logger.Printf("Failed to write response: %v", err)
	}
}

// registerTools is defined in tools.go — it registers all 16 MCP tools.
