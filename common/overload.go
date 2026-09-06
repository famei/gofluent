package common

// This file documents the migration of common/overload.py. The Python module
// provides `singledispatchmethod`, which dispatches an instance method on the
// runtime type of its first argument.
//
// Go has no runtime single-dispatch; the same effect is achieved with type
// switches. The constructors of Action (see icon.go) and the helper ToQIcon
// demonstrate the pattern:
//
//	switch v := value.(type) {
//	case string:
//		...
//	case FluentIconBase:
//		...
//	default:
//		...
//	}
//
// No runtime helper is required, so this package intentionally exports nothing.
