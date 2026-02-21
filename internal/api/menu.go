package api

import (
	"encoding/json"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/security"
)

// RegisterMenu registers application menu API handlers with security checks.
func RegisterMenu(router *ipc.Router, policy *security.Policy) {
	wrap := func(handler ipc.HandlerFunc) ipc.HandlerFunc {
		return func(params json.RawMessage) (any, error) {
			if err := policy.Check(security.PermMenu); err != nil {
				return nil, err
			}
			return handler(params)
		}
	}
	router.Handle("menu.set", wrap(handleMenuSet))
}
