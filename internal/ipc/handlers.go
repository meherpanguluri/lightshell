package ipc

// RegisterHandlers is called by the API packages to register their handlers.
// This file serves as the coordination point — each api package calls
// router.Handle() during setup.

// SetupRouter creates and returns a router with no handlers registered.
// API handlers are registered by calling router.Handle() from each api package.
func SetupRouter() *Router {
	return NewRouter()
}
