//go:build linux

package runtime

import "runtime"

func init() {
	// GTK requires the main thread
	runtime.LockOSThread()
}
