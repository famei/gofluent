//go:build !windows

package win32

import "unsafe"

// On non-Windows platforms the frameless-window effects are unsupported; every
// function is a no-op so the rest of the library keeps compiling.

func DefWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr { return 0 }

func GetWindowRect(hwnd HWND, rect *RECT) bool { return false }

func IsZoomed(hwnd HWND) bool { return false }

func GetSystemMetrics(index int) int { return 0 }

func SendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr { return 0 }

func ReleaseCapture() {}

// GetCursorPos is a no-op off Windows.
func GetCursorPos() (x, y int32) { return 0, 0 }

func GetWindowLongPtr(hwnd HWND, index int) uintptr { return 0 }

func SetWindowLongPtr(hwnd HWND, index int, newLong uintptr) uintptr { return 0 }

func CallWindowProcW(prevWndProc uintptr, hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	return 0
}

func DwmExtendFrameIntoClientArea(hwnd HWND, margins *MARGINS) error { return nil }

func DwmSetWindowAttribute(hwnd HWND, attr uint32, value unsafe.Pointer, size uintptr) error {
	return nil
}

func SetWindowCompositionAttribute(hwnd HWND, data *WINDOWCOMPOSITIONATTRIBDATA) error {
	return nil
}

func EnableMica(hwnd HWND, dark bool) {}

func EnableAcrylic(hwnd HWND, gradientColor uint32) {}

func RemoveBackdrop(hwnd HWND) {}

func EnableRoundedCorners(hwnd HWND) {}
