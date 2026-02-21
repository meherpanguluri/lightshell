package api

import (
	"encoding/json"

	"github.com/meherpanguluri/lightshell/internal/ipc"
	"github.com/meherpanguluri/lightshell/internal/security"
)

// RegisterNotification registers notification API handlers with security checks.
func RegisterNotification(router *ipc.Router, policy *security.Policy) {
	wrap := func(handler ipc.HandlerFunc) ipc.HandlerFunc {
		return func(params json.RawMessage) (any, error) {
			if err := policy.Check(security.PermNotification); err != nil {
				return nil, err
			}
			return handler(params)
		}
	}
	router.Handle("notify.send", wrap(handleNotifySend))
}
