#include "shim.h"

#include <QImage>
#include <QMouseEvent>
#include <QMoveEvent>
#include <QOpenGLWidget>
#include <QShowEvent>
#include <QWidget>

#include <cstdlib>
#include <cstring>

// AcrylicGLSurface is a non-native OpenGL surface. Qt renders it into an FBO and
// composites that through the normal widget pipeline, so ordinary child widgets
// can be placed on top of it.
class AcrylicGLSurface : public QOpenGLWidget
{
public:
    explicit AcrylicGLSurface(QWidget *parent) : QOpenGLWidget(parent) {}

    typedef void (*cb_void)(uintptr_t);
    typedef void (*cb_size)(uintptr_t, int, int);
    typedef void (*cb_mouse)(uintptr_t, int, int, int);
    typedef void (*cb_drag)(uintptr_t, int, int);

    uintptr_t ctx = 0;
    cb_void cbInit = nullptr;
    cb_void cbPaint = nullptr;
    cb_size cbResize = nullptr;
    cb_void cbShow = nullptr;
    cb_void cbMove = nullptr;
    cb_mouse cbPress = nullptr;
    cb_drag cbDrag = nullptr;
    cb_void cbRelease = nullptr;

protected:
    void initializeGL() override
    {
        if (cbInit) cbInit(ctx);
    }

    void paintGL() override
    {
        if (cbPaint) cbPaint(ctx);
    }

    void resizeGL(int width, int height) override
    {
        if (cbResize) cbResize(ctx, width, height);
    }

    void showEvent(QShowEvent *event) override
    {
        QOpenGLWidget::showEvent(event);
        if (cbShow) cbShow(ctx);
    }

    void moveEvent(QMoveEvent *event) override
    {
        QOpenGLWidget::moveEvent(event);
        if (cbMove) cbMove(ctx);
    }

    void mousePressEvent(QMouseEvent *event) override
    {
        QOpenGLWidget::mousePressEvent(event);
        if (cbPress) cbPress(ctx, static_cast<int>(event->button()), event->globalPos().x(), event->globalPos().y());
    }

    void mouseMoveEvent(QMouseEvent *event) override
    {
        QOpenGLWidget::mouseMoveEvent(event);
        if (cbDrag) cbDrag(ctx, event->globalPos().x(), event->globalPos().y());
    }

    void mouseReleaseEvent(QMouseEvent *event) override
    {
        QOpenGLWidget::mouseReleaseEvent(event);
        if (cbRelease) cbRelease(ctx);
    }
};

// Implemented in Go (see glsurface.go).
extern "C" {
void acrylicGLSurfaceInit(uintptr_t ctx);
void acrylicGLSurfacePaint(uintptr_t ctx);
void acrylicGLSurfaceResize(uintptr_t ctx, int width, int height);
void acrylicGLSurfaceShow(uintptr_t ctx);
void acrylicGLSurfaceMove(uintptr_t ctx);
void acrylicGLSurfacePress(uintptr_t ctx, int button, int globalX, int globalY);
void acrylicGLSurfaceDrag(uintptr_t ctx, int globalX, int globalY);
void acrylicGLSurfaceRelease(uintptr_t ctx);
}

extern "C" void *acrylic_qogl_new(void *parent)
{
    return new AcrylicGLSurface(static_cast<QWidget *>(parent));
}

extern "C" void acrylic_qogl_delete(void *w)
{
    delete static_cast<AcrylicGLSurface *>(w);
}

extern "C" void acrylic_qogl_use_go_callbacks(void *w, uintptr_t ctx)
{
    AcrylicGLSurface *s = static_cast<AcrylicGLSurface *>(w);
    s->ctx = ctx;
    s->cbInit = acrylicGLSurfaceInit;
    s->cbPaint = acrylicGLSurfacePaint;
    s->cbResize = acrylicGLSurfaceResize;
    s->cbShow = acrylicGLSurfaceShow;
    s->cbMove = acrylicGLSurfaceMove;
    s->cbPress = acrylicGLSurfacePress;
    s->cbDrag = acrylicGLSurfaceDrag;
    s->cbRelease = acrylicGLSurfaceRelease;
}

extern "C" void acrylic_qogl_set_geometry(void *w, int x, int y, int width, int height)
{
    static_cast<QWidget *>(w)->setGeometry(x, y, width, height);
}

extern "C" void acrylic_qogl_update(void *w)
{
    static_cast<QWidget *>(w)->update();
}

extern "C" int acrylic_qogl_is_valid(void *w)
{
    AcrylicGLSurface *s = static_cast<AcrylicGLSurface *>(w);
    return (s && s->isValid()) ? 1 : 0;
}

extern "C" unsigned char *acrylic_qogl_grab(void *w, int *width, int *height, int *stride)
{
    AcrylicGLSurface *s = static_cast<AcrylicGLSurface *>(w);
    if (!s) {
        return nullptr;
    }

    const QImage grabbed = s->grabFramebuffer();
    if (grabbed.isNull()) {
        return nullptr;
    }

    // RGBA8888 has the byte order Go expects, and matches miqt's Format_RGBA8888.
    const QImage rgba = grabbed.convertToFormat(QImage::Format_RGBA8888);
    const int bytes = rgba.bytesPerLine() * rgba.height();
    unsigned char *out = static_cast<unsigned char *>(malloc(static_cast<size_t>(bytes)));
    if (!out) {
        return nullptr;
    }
    memcpy(out, rgba.constBits(), static_cast<size_t>(bytes));

    *width = rgba.width();
    *height = rgba.height();
    *stride = rgba.bytesPerLine();
    return out;
}

extern "C" void acrylic_grab_free(unsigned char *pixels)
{
    free(pixels);
}
