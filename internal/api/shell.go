package api

import (
	"encoding/json"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/security"
)

// RegisterShell registers shell API handlers with security checks.
func RegisterShell(router *ipc.Router, policy *security.Policy) {
	wrap := func(handler ipc.HandlerFunc) ipc.HandlerFunc {
		return func(params json.RawMessage) (any, error) {
			if err := policy.Check(security.PermShell); err != nil {
				return nil, err
			}
			return handler(params)
		}
	}
	router.Handle("shell.open", wrap(handleShellOpen))
}
