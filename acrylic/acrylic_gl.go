package acrylic

import (
	"fmt"
	"os"
	"unsafe"

	gl "github.com/go-gl/gl/v2.1/gl"
	"github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/opengl"
)

func glLogf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[acrylic-gl] "+format+"\n", args...)
}

// AcrylicGLWidget is a QGLWidget that performs the acrylic blur on the GPU.
//
// It shares the backdrop-capture strategy of AcrylicWidget but uploads the
// captured pixels as an OpenGL texture and blurs them in a GLSL fragment
// shader instead of on the CPU.
type AcrylicGLWidget struct {
	*opengl.QGLWidget

	cached       *qt.QImage // downscaled backdrop (unflipped), uploaded for the shader blur
	sharp        *qt.QImage // full-resolution backdrop, sampled outside rounded corners
	composite    *qt.QImage // CPU-blurred copy: how the card looks when another card blurs it
	textureDirty bool       // cached/sharp changed since the last upload

	caption string

	downsampleFactor    int
	blurRadius          int // shader kernel radius (0..8)
	blurSigma           float64
	noiseAmount         int
	tintR, tintG, tintB int     // 0..255
	tintOpacity         float64 // 0..1

	cornerRadius                       int
	borderWidth                        int
	hasBorder                          bool
	borderR, borderG, borderB, borderA int

	program      uint32
	texture      uint32
	sharpTexture uint32
	uTex         int32
	uSharp       int32
	uTexel       int32
	uTint        int32
	uRadius      int32
	uSigma       int32
	uSize        int32
	uCorner      int32
	uBorder      int32
	uBorderW     int32
	uNoise       int32
	texWidth     int
	texHeight    int
	texelX       float32
	texelY       float32

	glInitialized bool

	capturing        bool
	refreshTimer     *qt.QTimer
	refreshScheduled bool

	draggable                          bool
	dragging                           bool
	dragPressGlobalX, dragPressGlobalY int
	dragPressPosX, dragPressPosY       int
	dragCache                          *qt.QImage // window composite, reused during one drag

	onUpdated     func() // called after every finished capture (see OnUpdated)
	silentRefresh bool   // refresh asked for from inside a callback: do not notify again
}

// NewAcrylicGLWidget creates an AcrylicGLWidget with sensible defaults.
func NewAcrylicGLWidget(parent *qt.QWidget) *AcrylicGLWidget {
	w := &AcrylicGLWidget{
		QGLWidget:        opengl.NewQGLWidget(parent),
		downsampleFactor: 4,
		blurRadius:       3,
		blurSigma:        1.7,
		noiseAmount:      2,
		tintR:            255,
		tintG:            255,
		tintB:            255,
		tintOpacity:      0.30,
		uTex:             -1,
		uSharp:           -1,
		uTexel:           -1,
		uTint:            -1,
		uRadius:          -1,
		uSigma:           -1,
		uSize:            -1,
		uCorner:          -1,
		uBorder:          -1,
		uBorderW:         -1,
		uNoise:           -1,
	}

	w.refreshTimer = qt.NewQTimer2(w.QGLWidget.QObject)
	w.refreshTimer.SetSingleShot(true)
	w.refreshTimer.OnTimeout(func() {
		w.refreshScheduled = false
		w.doCapture()
	})

	w.OnInitializeGL(func(super func()) {
		w.initGL()
	})
	w.OnResizeGL(func(super func(width int, height int), width int, height int) {
		gl.Viewport(0, 0, int32(width), int32(height))
	})
	w.OnPaintGL(func(super func()) {
		w.paintGL()
	})

	w.OnShowEvent(func(super func(ev *qt.QShowEvent), ev *qt.QShowEvent) {
		super(ev)
		if captureVisibilityGuarded() {
			return // re-shown by an in-progress capture; it is up to date already
		}
		w.scheduleRefresh()
	})
	w.OnResizeEvent(func(super func(ev *qt.QResizeEvent), ev *qt.QResizeEvent) {
		super(ev)
		w.scheduleRefresh()
	})
	w.OnMoveEvent(func(super func(ev *qt.QMoveEvent), ev *qt.QMoveEvent) {
		super(ev)
		// Off-screen sibling compositing: the backdrop follows the card live.
		w.scheduleRefresh()
	})

	w.OnMousePressEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		super(ev)
		if !w.draggable || ev.Button() != qt.LeftButton {
			return
		}
		w.dragging = true
		w.dragCache = nil
		w.dragPressGlobalX = ev.GlobalX()
		w.dragPressGlobalY = ev.GlobalY()
		pos := w.Pos()
		w.dragPressPosX = pos.X()
		w.dragPressPosY = pos.Y()
		w.Raise()
		moveCardToTop(w)
		// One-off z-order invalidation; see acrylic.go / refreshCardsBelow.
		refreshCardsBelow(w)
	})
	w.OnMouseMoveEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		super(ev)
		if !w.dragging {
			return
		}
		nx := w.dragPressPosX + (ev.GlobalX() - w.dragPressGlobalX)
		ny := w.dragPressPosY + (ev.GlobalY() - w.dragPressGlobalY)
		w.Move(nx, ny)
	})
	w.OnMouseReleaseEvent(func(super func(ev *qt.QMouseEvent), ev *qt.QMouseEvent) {
		super(ev)
		if !w.dragging {
			return
		}
		w.dragging = false
		w.dragCache = nil
		w.scheduleRefresh()
	})

	registerCard(w)
	return w
}

// Delete releases the underlying Qt objects.
func (w *AcrylicGLWidget) Delete() {
	unregisterCard(w)
	w.refreshTimer.Stop()
	w.QGLWidget.Delete()
}

// Refresh forces a re-capture of the backdrop on the next event-loop tick.
//
// Calling it from inside an OnUpdated callback re-captures *without* firing this
// card's callback again, so the usual wiring (every card refreshing the others)
// cannot loop.
func (w *AcrylicGLWidget) Refresh() {
	if insideNotify() {
		w.silentRefresh = true
	}
	w.scheduleRefresh()
}

// BackdropImage returns the last captured backdrop (mainly for tests).
func (w *AcrylicGLWidget) BackdropImage() *qt.QImage { return w.sharp }

// SetTintColor sets the overlay tint colour (0..1 per channel).
func (w *AcrylicGLWidget) SetTintColor(hex string) {
	if r, g, b, a, ok := parseHexColor(hex); ok {
		w.tintR, w.tintG, w.tintB = r, g, b
		w.tintOpacity = float64(a) / 255.0
		w.UpdateGL()
	}
}

// SetTintRGB sets the overlay tint colour numerically (0..255 per channel).
func (w *AcrylicGLWidget) SetTintRGB(r, g, b int) {
	w.tintR, w.tintG, w.tintB = clamp255(r), clamp255(g), clamp255(b)
	w.UpdateGL()
}

// SetTintOpacity sets the tint strength (0 = none, 1 = fully opaque tint).
func (w *AcrylicGLWidget) SetTintOpacity(opacity float64) {
	w.tintOpacity = opacity
	w.UpdateGL()
}

// SetCornerRadius rounds the card's corners (radius in pixels, 0 = square).
func (w *AcrylicGLWidget) SetCornerRadius(radius int) {
	if radius < 0 {
		radius = 0
	}
	w.cornerRadius = radius
	w.UpdateGL()
}

// SetBorderColor sets the border colour from a hex string, e.g. "#ffffff80".
func (w *AcrylicGLWidget) SetBorderColor(hex string) {
	if r, g, b, a, ok := parseHexColor(hex); ok {
		w.borderR, w.borderG, w.borderB, w.borderA = r, g, b, a
		w.hasBorder = true
		w.UpdateGL()
	}
}

// SetBorderRGB sets the border colour numerically (0..255 per channel).
func (w *AcrylicGLWidget) SetBorderRGB(r, g, b, a int) {
	w.borderR, w.borderG, w.borderB, w.borderA = clamp255(r), clamp255(g), clamp255(b), clamp255(a)
	w.hasBorder = true
	w.UpdateGL()
}

// SetBorderWidth sets the border width in pixels (0 disables the border).
func (w *AcrylicGLWidget) SetBorderWidth(width int) {
	if width < 0 {
		width = 0
	}
	w.borderWidth = width
	w.UpdateGL()
}

// SetBlurRadius sets the shader kernel radius (0..8).
func (w *AcrylicGLWidget) SetBlurRadius(radius int) {
	if radius < 0 {
		radius = 0
	}
	if radius > 8 {
		radius = 8
	}
	w.blurRadius = radius
	w.UpdateGL()
}

// SetDownsampleFactor sets the downscale factor applied before uploading.
func (w *AcrylicGLWidget) SetDownsampleFactor(factor int) {
	if factor < 1 {
		factor = 1
	}
	w.downsampleFactor = factor
	w.Refresh()
}

// SetNoiseAmount sets the shader grain strength (0 disables it).
func (w *AcrylicGLWidget) SetNoiseAmount(amount int) {
	if amount < 0 {
		amount = 0
	}
	w.noiseAmount = amount
	w.UpdateGL()
}

// SetDraggable enables mouse dragging of the widget; the backdrop follows the
// card live while it is dragged.
func (w *AcrylicGLWidget) SetDraggable(draggable bool) {
	w.draggable = draggable
}

// SetCaption sets an optional label rendered in the top-left corner.
func (w *AcrylicGLWidget) SetCaption(text string) {
	w.caption = text
	w.UpdateGL()
}

func (w *AcrylicGLWidget) scheduleRefresh() {
	if w.capturing || w.refreshScheduled || captureVisibilityGuarded() || refreshingSelf(w) {
		return
	}
	w.refreshScheduled = true
	w.refreshTimer.Start(0)
}

func (w *AcrylicGLWidget) doCapture() {
	if w.capturing {
		return
	}
	w.capturing = true
	defer func() { w.capturing = false }()

	img := captureBackdropFor(w, w.backdropCachePtr())
	if img == nil {
		return
	}
	w.sharp = img
	w.cached = downsampleForGL(img, w.downsampleFactor)
	// The shader blurs w.cached on the GPU. When this card is composited into a
	// sibling's backdrop there is no shader involved, so keep a CPU-blurred copy
	// whose appearance matches what the shader shows on screen. Otherwise a
	// sibling's rounded corner would reveal this card's unblurred backdrop and
	// look like a washed-out acrylic.
	w.composite = downsampleAndBlur(img, w.downsampleFactor, w.blurRadius, w.noiseAmount, w.blurSigma)
	w.textureDirty = true
	w.UpdateGL()

	// Cards above blur this card, so their backdrops just went stale.
	refreshCardsAbove(w)
	if w.silentRefresh {
		w.silentRefresh = false // follow-up refresh: the callbacks already ran
	} else {
		w.notifyUpdated()
	}
}

// OnUpdated registers a callback that runs after this card has finished updating
// its backdrop, i.e. whenever what the card shows has changed.
//
// It is the hook for keeping *other* acrylic widgets in sync: cards stacked above
// this one in the same window are refreshed automatically, but cards in another
// window (or any other widget that shows the same backdrop) are not. The callback
// runs on the GUI thread right after the card is up to date, and it may fire once
// per capture — keep it cheap, e.g. call Refresh() on the other widget. Pass nil to
// remove it.
func (w *AcrylicGLWidget) OnUpdated(fn func()) {
	w.onUpdated = fn
}

func (w *AcrylicGLWidget) notifyUpdated() {
	notifyCardUpdated(w, w.onUpdated)
}

// backdropCachePtr returns the per-drag composite cache while the card is being
// dragged, else nil. Reusing it keeps the acrylic content glued to the card
// while dragging instead of lagging behind it (no smear).
func (w *AcrylicGLWidget) backdropCachePtr() **qt.QImage {
	if w.dragging {
		return &w.dragCache
	}
	return nil
}

// widget returns the underlying QWidget (used for sibling classification).
func (w *AcrylicGLWidget) widget() *qt.QWidget { return w.QGLWidget.QWidget }

// qtPaintsCard reports that Qt does NOT paint this card as ordinary child
// content: it is a native window, and Qt leaves the area of native children
// unpainted when a parent renders its subtree. Hiding it for a capture would only
// make the native window flicker.
func (w *AcrylicGLWidget) qtPaintsCard() bool { return false }

// drawAcrylic paints this card's acrylic content (rounded, downscaled backdrop
// + tint + border) at the given offset for sibling compositing.
func (w *AcrylicGLWidget) drawAcrylic(painter *qt.QPainter, x, y int) {
	width := w.Width()
	height := w.Height()

	painter.Save()
	painter.SetRenderHint2(qt.QPainter__Antialiasing, true)
	painter.SetRenderHint2(qt.QPainter__SmoothPixmapTransform, true)

	if w.cornerRadius > 0 {
		path := qt.NewQPainterPath()
		path.AddRoundedRect2(float64(x), float64(y), float64(width), float64(height),
			float64(w.cornerRadius), float64(w.cornerRadius))
		painter.SetClipPath(path)
		path.Delete()
	}

	// Prefer the CPU-blurred copy: it reproduces this card's on-screen shader
	// output, so a sibling compositing it sees the same acrylic the user sees.
	source := w.composite
	if source == nil {
		source = w.cached
	}
	if source != nil {
		rect := qt.NewQRect4(x, y, width, height)
		painter.DrawImage6(rect, source)
		rect.Delete()
	} else {
		fill := qt.NewQColor11(w.tintR, w.tintG, w.tintB, 255)
		painter.FillRect5(x, y, width, height, fill)
		fill.Delete()
	}

	if w.tintOpacity > 0 {
		a := clamp255(int(w.tintOpacity * 255))
		tint := qt.NewQColor11(w.tintR, w.tintG, w.tintB, a)
		painter.FillRect5(x, y, width, height, tint)
		tint.Delete()
	}

	// Child widgets sit on top of the acrylic, so they belong in the composite.
	renderChildWidgets(w.QGLWidget.QWidget, painter, x, y)

	painter.Restore()

	drawShaderStyleBorder(painter, x, y, width, height,
		w.hasBorder, w.borderWidth, w.cornerRadius,
		w.borderR, w.borderG, w.borderB, w.borderA)
}

// downsampleForGL downscales the captured image (the flip is applied at upload
// time, since QPainter compositing needs the unflipped orientation).
func downsampleForGL(img *qt.QImage, factor int) *qt.QImage {
	sw := img.Width() / factor
	sh := img.Height() / factor
	if sw < 1 {
		sw = 1
	}
	if sh < 1 {
		sh = 1
	}

	small := img.Scaled3(sw, sh, qt.IgnoreAspectRatio, qt.SmoothTransformation)
	if small.Format() != qt.QImage__Format_RGBA8888 {
		small = small.ConvertToFormat(qt.QImage__Format_RGBA8888)
	}
	return small
}

func (w *AcrylicGLWidget) initGL() {
	if err := gl.Init(); err != nil {
		// OpenGL 2.1 (or better) is required; leave the widget blank on failure.
		glLogf("gl.Init failed: %v", err)
		return
	}
	w.program = buildProgram()
	if w.program == 0 {
		glLogf("program build failed")
		return
	}

	w.uTex = gl.GetUniformLocation(w.program, gl.Str("uTex\x00"))
	w.uSharp = gl.GetUniformLocation(w.program, gl.Str("uSharp\x00"))
	w.uTexel = gl.GetUniformLocation(w.program, gl.Str("uTexel\x00"))
	w.uTint = gl.GetUniformLocation(w.program, gl.Str("uTint\x00"))
	w.uRadius = gl.GetUniformLocation(w.program, gl.Str("uRadius\x00"))
	w.uSigma = gl.GetUniformLocation(w.program, gl.Str("uSigma\x00"))
	w.uSize = gl.GetUniformLocation(w.program, gl.Str("uSize\x00"))
	w.uCorner = gl.GetUniformLocation(w.program, gl.Str("uCorner\x00"))
	w.uBorder = gl.GetUniformLocation(w.program, gl.Str("uBorder\x00"))
	w.uBorderW = gl.GetUniformLocation(w.program, gl.Str("uBorderW\x00"))
	w.uNoise = gl.GetUniformLocation(w.program, gl.Str("uNoise\x00"))

	gl.UseProgram(w.program)
	if w.uTex >= 0 {
		gl.Uniform1i(w.uTex, 0)
	}
	if w.uSharp >= 0 {
		gl.Uniform1i(w.uSharp, 1)
	}
	w.glInitialized = true
}

func (w *AcrylicGLWidget) paintGL() {
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	if w.glInitialized {
		if w.textureDirty {
			w.uploadTextures()
			w.textureDirty = false
		}
		if w.texture != 0 && w.sharpTexture != 0 {
			w.drawAcrylicQuad()
		}
	}

	// The caption is drawn with QGLWidget's own text rendering, which renders the
	// string offscreen and blits it as a texture. A QPainter on this widget would
	// disturb the native GL surface — the card then showed the widget's palette
	// colour instead of the acrylic.
	if w.caption != "" {
		font := captionFont(w.QGLWidget.QWidget)
		colour := qt.NewQColor11(255, 255, 255, 255)
		w.QglColor(colour)
		w.RenderText3(16, 30, w.caption, font)
		colour.Delete()
		font.Delete()
	}
}

// drawAcrylicQuad binds the two textures, sets every uniform and draws the
// full-size quad. It must run inside beginNativePainting.
func (w *AcrylicGLWidget) drawAcrylicQuad() {
	gl.UseProgram(w.program)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, w.texture)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, w.sharpTexture)
	gl.ActiveTexture(gl.TEXTURE0)

	if w.uTexel >= 0 {
		gl.Uniform2f(w.uTexel, w.texelX, w.texelY)
	}
	if w.uTint >= 0 {
		gl.Uniform4f(w.uTint,
			float32(w.tintR)/255.0, float32(w.tintG)/255.0, float32(w.tintB)/255.0,
			float32(w.tintOpacity))
	}
	if w.uRadius >= 0 {
		gl.Uniform1i(w.uRadius, int32(w.blurRadius))
	}
	if w.uSigma >= 0 {
		gl.Uniform1f(w.uSigma, float32(w.blurSigma))
	}
	if w.uSize >= 0 {
		gl.Uniform2f(w.uSize, float32(w.Width()), float32(w.Height()))
	}
	if w.uCorner >= 0 {
		gl.Uniform1f(w.uCorner, float32(w.cornerRadius))
	}
	if w.uBorder >= 0 {
		gl.Uniform4f(w.uBorder,
			float32(w.borderR)/255.0, float32(w.borderG)/255.0, float32(w.borderB)/255.0,
			float32(w.borderA)/255.0)
	}
	if w.uBorderW >= 0 {
		gl.Uniform1f(w.uBorderW, float32(w.borderWidth))
	}
	if w.uNoise >= 0 {
		gl.Uniform1f(w.uNoise, float32(w.noiseAmount))
	}

	drawQuad()
}

// uploadTextures uploads both the downsampled blur texture (unit 0) and the
// full-resolution sharp texture used outside the rounded corners (unit 1).
func (w *AcrylicGLWidget) uploadTextures() {
	if w.cached != nil {
		// OpenGL textures are bottom-up, so flip vertically.
		flipped := w.cached.Mirrored2(false, true)
		w.texWidth = flipped.Width()
		w.texHeight = flipped.Height()

		if w.texture == 0 {
			gl.GenTextures(1, &w.texture)
		}
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, w.texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(w.texWidth), int32(w.texHeight),
			0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(flipped.Bits()))

		w.texelX = 1.0 / float32(w.texWidth)
		w.texelY = 1.0 / float32(w.texHeight)
	}

	if w.sharp != nil {
		sflipped := w.sharp.Mirrored2(false, true)

		if w.sharpTexture == 0 {
			gl.GenTextures(1, &w.sharpTexture)
		}
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, w.sharpTexture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(sflipped.Width()), int32(sflipped.Height()),
			0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(sflipped.Bits()))
	}

	gl.ActiveTexture(gl.TEXTURE0)

	// Restore the packing state we changed: Qt's own painter uploads (the glyph
	// atlas for the caption, for instance) go through the same context, and a
	// leaked UNPACK_ALIGNMENT of 1 makes the glyph rows line up differently.
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func buildProgram() uint32 {
	vs := compileShader(gl.VERTEX_SHADER, vertexShaderSource)
	if vs == 0 {
		return 0
	}
	fs := compileShader(gl.FRAGMENT_SHADER, fragmentShaderSource)
	if fs == 0 {
		gl.DeleteShader(vs)
		return 0
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)
	gl.LinkProgram(program)

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == 0 {
		var logLen int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLen)
		if logLen > 1 {
			log := make([]byte, logLen)
			gl.GetProgramInfoLog(program, logLen, nil, &log[0])
			glLogf("program link failed: %s", string(log))
		} else {
			glLogf("program link failed (no log)")
		}
		gl.DeleteProgram(program)
		return 0
	}
	return program
}

func compileShader(xtype uint32, source string) uint32 {
	shader := gl.CreateShader(xtype)
	csrc, free := gl.Strs(source)
	defer free()

	length := int32(len(source))
	gl.ShaderSource(shader, 1, csrc, &length)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == 0 {
		var logLen int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLen)
		if logLen > 1 {
			log := make([]byte, logLen)
			gl.GetShaderInfoLog(shader, logLen, nil, &log[0])
			glLogf("shader compile failed: %s", string(log))
		} else {
			glLogf("shader compile failed (no log)")
		}
		gl.DeleteShader(shader)
		return 0
	}
	return shader
}

// drawQuad draws a full-screen textured quad using the fixed-function vertex
// pipeline (gl_MultiTexCoord0 is consumed by the vertex shader). The texture
// was already flipped vertically on upload, so texture v maps directly to
// screen y.
func drawQuad() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, 1, 0, 1, -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(0, 0)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(1, 0)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(1, 1)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(0, 1)
	gl.End()
}

