// Package glsurface wraps a non-native QOpenGLWidget through a small cgo shim.
//
// QGLWidget is a native window: Qt cannot composite ordinary child widgets over
// its GL content, and everything a child widget leaves unpainted shows the bare
// (black) surface. QOpenGLWidget renders into an FBO that Qt composites through
// the normal widget pipeline, so children behave exactly as they do on any other
// widget. miqt has no bindings for it, hence this minimal shim.
//
// miqt can only install virtual-method overrides on objects it constructed
// itself, so the widget events this package needs are reimplemented in C++ and
// forwarded to the Go callbacks below.
package glsurface

/*
#include <stdint.h>
#include <stdlib.h>
#include "shim.h"
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// Handlers are the callbacks a Surface invokes. All of them run on the GUI
// thread; the GL ones run with the surface's context current.
type Handlers struct {
	Init    func()
	Paint   func()
	Resize  func(width, height int)
	Show    func()
	Move    func()
	Press   func(button, globalX, globalY int)
	Drag    func(globalX, globalY int)
	Release func()
}

// Surface is one non-native OpenGL surface widget.
type Surface struct {
	ptr unsafe.Pointer

	handle cgo.Handle

	handlers Handlers
}

// New creates a surface as a child of the given parent widget (a raw QWidget
// pointer, as returned by qt.QWidget.UnsafePointer), then installs the callbacks.
func New(parent unsafe.Pointer, handlers Handlers) *Surface {
	s := &Surface{handlers: handlers}
	s.handle = cgo.NewHandle(s)
	s.ptr = C.acrylic_qogl_new(parent)
	C.acrylic_qogl_use_go_callbacks(s.ptr, C.uintptr_t(s.handle))
	return s
}

// Pointer returns the raw QWidget pointer, so the caller can wrap the same
// object as a *qt.QWidget.
func (s *Surface) Pointer() unsafe.Pointer { return s.ptr }

// SetGeometry moves and resizes the surface.
func (s *Surface) SetGeometry(x, y, width, height int) {
	C.acrylic_qogl_set_geometry(s.ptr, C.int(x), C.int(y), C.int(width), C.int(height))
}

// Update schedules a repaint of the surface.
func (s *Surface) Update() {
	C.acrylic_qogl_update(s.ptr)
}

// IsValid reports whether the surface has a usable OpenGL context.
func (s *Surface) IsValid() bool {
	return C.acrylic_qogl_is_valid(s.ptr) != 0
}

// GrabFrameBuffer renders the surface into its framebuffer and returns the
// pixels as RGBA8888, together with the width, height and row stride in bytes.
func (s *Surface) GrabFrameBuffer() (pixels []byte, width, height, stride int) {
	var cw, ch, cstride C.int
	buf := C.acrylic_qogl_grab(s.ptr, &cw, &ch, &cstride)
	if buf == nil {
		return nil, 0, 0, 0
	}
	defer C.acrylic_grab_free(buf)

	width, height, stride = int(cw), int(ch), int(cstride)
	out := make([]byte, stride*height)
	if len(out) > 0 {
		copy(out, unsafe.Slice((*byte)(unsafe.Pointer(buf)), len(out)))
	}
	return out, width, height, stride
}

// Delete destroys the underlying Qt widget.
func (s *Surface) Delete() {
	if s.ptr == nil {
		return
	}
	C.acrylic_qogl_delete(s.ptr)
	s.ptr = nil
	s.handle.Delete()
}

// surfaceFor resolves the Go surface a C callback refers to.
func surfaceFor(ctx C.uintptr_t) *Surface {
	s, ok := cgo.Handle(ctx).Value().(*Surface)
	if !ok {
		return nil
	}
	return s
}

//export acrylicGLSurfaceInit
func acrylicGLSurfaceInit(ctx C.uintptr_t) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Init != nil {
		s.handlers.Init()
	}
}

//export acrylicGLSurfacePaint
func acrylicGLSurfacePaint(ctx C.uintptr_t) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Paint != nil {
		s.handlers.Paint()
	}
}

//export acrylicGLSurfaceResize
func acrylicGLSurfaceResize(ctx C.uintptr_t, width, height C.int) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Resize != nil {
		s.handlers.Resize(int(width), int(height))
	}
}

//export acrylicGLSurfaceShow
func acrylicGLSurfaceShow(ctx C.uintptr_t) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Show != nil {
		s.handlers.Show()
	}
}

//export acrylicGLSurfaceMove
func acrylicGLSurfaceMove(ctx C.uintptr_t) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Move != nil {
		s.handlers.Move()
	}
}

//export acrylicGLSurfacePress
func acrylicGLSurfacePress(ctx C.uintptr_t, button, globalX, globalY C.int) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Press != nil {
		s.handlers.Press(int(button), int(globalX), int(globalY))
	}
}

//export acrylicGLSurfaceDrag
func acrylicGLSurfaceDrag(ctx C.uintptr_t, globalX, globalY C.int) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Drag != nil {
		s.handlers.Drag(int(globalX), int(globalY))
	}
}

//export acrylicGLSurfaceRelease
func acrylicGLSurfaceRelease(ctx C.uintptr_t) {
	if s := surfaceFor(ctx); s != nil && s.handlers.Release != nil {
		s.handlers.Release()
	}
}
