package api

import (
	"encoding/json"

	"github.com/meherpanguluri/lightshell/internal/ipc"
	"github.com/meherpanguluri/lightshell/internal/security"
)

// RegisterClipboard registers clipboard API handlers with security checks.
func RegisterClipboard(router *ipc.Router, policy *security.Policy) {
	wrap := func(handler ipc.HandlerFunc) ipc.HandlerFunc {
		return func(params json.RawMessage) (any, error) {
			if err := policy.Check(security.PermClipboard); err != nil {
				return nil, err
			}
			return handler(params)
		}
	}
	router.Handle("clipboard.read", wrap(handleClipboardRead))
	router.Handle("clipboard.write", wrap(handleClipboardWrite))
}
