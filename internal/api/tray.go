package api

import (
	"encoding/json"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/security"
)

// RegisterTray registers system tray API handlers with security checks.
func RegisterTray(router *ipc.Router, policy *security.Policy) {
	wrap := func(handler ipc.HandlerFunc) ipc.HandlerFunc {
		return func(params json.RawMessage) (any, error) {
			if err := policy.Check(security.PermTray); err != nil {
				return nil, err
			}
			return handler(params)
		}
	}
	router.Handle("tray.set", wrap(handleTraySet))
	router.Handle("tray.remove", wrap(handleTrayRemove))
}
