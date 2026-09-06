package common

// ExceptionHandler mirrors the Python `exceptionHandler(*default)` decorator.
// In Go, defer/recover is idiomatic; this helper lets callers recover from a
// panic and return a fallback value of the same type.
//
//	fn := func() *qt.QColor { ... may panic ... }
//	v := ExceptionHandler(fn, defaultColor)()
func ExceptionHandler[T any](fn func() T, def T) func() T {
	return func() (out T) {
		defer func() {
			if recover() != nil {
				out = def
			}
		}()
		return fn()
	}
}

// SafeCall runs fn and returns (result, false) if it panics, otherwise
// (result, true).
func SafeCall[T any](fn func() T) (out T, ok bool) {
	ok = true
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	out = fn()
	return
}
