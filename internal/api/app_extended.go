package api

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/lightshell-dev/lightshell/internal/ipc"
)

// RegisterAppExtended registers extended app API handlers.
// These include badge count, second instance detection, and protocol handling.
func RegisterAppExtended(router *ipc.Router, appName string) {
	router.Handle("app.setBadgeCount", handleAppSetBadgeCount)

	// Second instance detection via Unix domain socket lockfile
	router.Handle("app.enableSingleInstance", func(params json.RawMessage) (any, error) {
		sockPath := singleInstanceSocketPath(appName)
		// Try connecting to existing socket — if successful, another instance is running
		conn, err := net.Dial("unix", sockPath)
		if err == nil {
			// Another instance is running — send our args and exit
			args := strings.Join(os.Args[1:], "\n")
			conn.Write([]byte(args))
			conn.Close()
			return map[string]bool{"isSecondInstance": true}, nil
		}

		// No existing instance — create the socket
		os.Remove(sockPath) // clean up stale socket
		listener, err := net.Listen("unix", sockPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create single instance lock: %w", err)
		}
		// Restrict socket to owner-only access
		os.Chmod(sockPath, 0600)

		// Listen for connections from second instances in the background
		go func() {
			defer listener.Close()
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				buf := make([]byte, 4096)
				n, _ := conn.Read(buf)
				conn.Close()
				if n > 0 {
					args := strings.Split(string(buf[:n]), "\n")
					router.SendEvent("app.secondInstance", map[string]any{
						"args": args,
					})
				}
			}
		}()

		return map[string]bool{"isSecondInstance": false}, nil
	})
}

func singleInstanceSocketPath(appName string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("lightshell-%s.sock", appName))
}
