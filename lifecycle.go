package bind

// Startup is implemented by items that need initialization after dependency injection.
// Startup() is called automatically when the item is added to the binder.
type Startup interface {
	Startup() error
}

// Shutdown is implemented by items that need cleanup during shutdown.
// Shutdown() is called automatically in reverse order when binder.Shutdown() is called.
type Shutdown interface {
	Shutdown() error
}
