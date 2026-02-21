//go:build darwin

package runtime

import "runtime"

func init() {
	// macOS requires the main thread for Cocoa operations
	runtime.LockOSThread()
}
