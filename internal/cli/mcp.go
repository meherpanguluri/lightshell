package cli

import (
	"os"

	"github.com/lightshell-dev/lightshell/internal/mcp"
)

// MCP starts the MCP (Model Context Protocol) server over stdio.
// The server speaks JSON-RPC 2.0 and exposes LightShell tools and resources
// for AI-assisted development.
func MCP() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// The API docs can be loaded from disk or embedded later.
	// For now, pass empty string to use the built-in default docs.
	server := mcp.NewServer(dir, "")
	return server.Run()
}
