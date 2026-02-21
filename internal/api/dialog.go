package api

import (
	"encoding/json"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/security"
)

// RegisterDialog registers dialog API handlers with security checks.
func RegisterDialog(router *ipc.Router, policy *security.Policy) {
	wrap := func(handler ipc.HandlerFunc) ipc.HandlerFunc {
		return func(params json.RawMessage) (any, error) {
			if err := policy.Check(security.PermDialog); err != nil {
				return nil, err
			}
			return handler(params)
		}
	}
	router.Handle("dialog.open", wrap(handleDialogOpen))
	router.Handle("dialog.save", wrap(handleDialogSave))
	router.Handle("dialog.message", wrap(handleDialogMessage))
	router.Handle("dialog.confirm", wrap(handleDialogConfirm))
	router.Handle("dialog.prompt", wrap(handleDialogPrompt))
}
