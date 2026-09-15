//go:build windowsqtstatic
// +build windowsqtstatic

package glsurface

// Static Qt5 needs the --static pkg-config flags, exactly like miqt's own
// bindings; otherwise the platform plugin fails to initialise at runtime.

/*
#cgo CXXFLAGS: -std=c++11
#cgo CFLAGS: -std=gnu11
#cgo pkg-config: --static Qt5Widgets
*/
import "C"
