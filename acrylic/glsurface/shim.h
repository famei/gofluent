#ifndef ACRYLIC_GLSURFACE_SHIM_H
#define ACRYLIC_GLSURFACE_SHIM_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Creates a non-native QOpenGLWidget. Being non-native, Qt composites its
// framebuffer through the ordinary widget pipeline, so normal child widgets can
// live on top of the GL content (unlike QGLWidget).
//
// miqt cannot install virtual-method overrides on an object it did not
// construct, so the events this widget needs are forwarded through these
// callbacks, which are implemented in Go (see glsurface.go).
void *acrylic_qogl_new(void *parent);
void acrylic_qogl_delete(void *w);
void acrylic_qogl_use_go_callbacks(void *w, uintptr_t ctx);
void acrylic_qogl_set_geometry(void *w, int x, int y, int width, int height);
void acrylic_qogl_update(void *w);
int acrylic_qogl_is_valid(void *w);

// Renders the surface into an offscreen buffer and copies it out as RGBA8888.
// Returns a malloc'd buffer the caller releases with acrylic_grab_free.
unsigned char *acrylic_qogl_grab(void *w, int *width, int *height, int *stride);
void acrylic_grab_free(unsigned char *pixels);

#ifdef __cplusplus
}
#endif

#endif
