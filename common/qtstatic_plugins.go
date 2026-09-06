//go:build windowsqtstatic

package common

/*
#cgo CXXFLAGS: -DMIQT_WINDOWSQTSTATIC
#cgo pkg-config: --static Qt5Gui
*/
import "C"
