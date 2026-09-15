package acrylic

import (
	"unsafe"

	gl "github.com/go-gl/gl/v2.1/gl"
	"github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/acrylic/glsurface"
)

// AcrylicOpenGLWidget is the GPU acrylic card built on a *non-native* OpenGL
// surface (QOpenGLWidget through glsurface). It blurs on the GPU exactly like
// AcrylicGLWidget and exposes the same API, but because Qt composites it through
// the ordinary widget pipeline:
//
//   - ordinary child widgets can be placed on top of it (a QGLWidget is a native
//     window, so children there leave black holes wherever they do not paint);
//   - its backdrop is captured with the same off-screen window composite as the
//     CPU card, so rounded corners show live content and sibling cards composite
//     faithfully;
//   - GrabFrameBuffer returns the card's real rendered pixels.
//
// Use AcrylicGLWidget only when the extra native window is acceptable and the
// card never hosts child widgets.
type AcrylicOpenGLWidget struct {
	*qt.QWidget

	surface *glsurface.Surface

	cached       *qt.QImage // downscaled backdrop (unflipped), uploaded for the shader blur
	sharp        *qt.QImage // full-resolution backdrop, sampled outside rounded corners
	composite    *qt.QImage // CPU-blurred copy: how the card looks when another card blurs it
	textureDirty bool

	caption      string
	captionLabel *qt.QLabel // real label child, so the text is crisp (see SetCaption)

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

	program       uint32
	texture       uint32
	sharpTexture  uint32
	uTex          int32
	uSharp        int32
	uTexel        int32
	uTint         int32
	uRadius       int32
	uSigma        int32
	uSize         int32
	uCorner       int32
	uBorder       int32
	uBorderW      int32
	uNoise        int32
	texWidth      int
	texHeight     int
	texelX        float32
	texelY        float32
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

// NewAcrylicOpenGLWidget creates a GPU acrylic card with sensible defaults. Its
// API matches NewAcrylicWidget and NewAcrylicGLWidget, so the same configuration
// code drives any of them.
func NewAcrylicOpenGLWidget(parent *qt.QWidget) *AcrylicOpenGLWidget {
	w := &AcrylicOpenGLWidget{
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

	var parentPtr unsafe.Pointer
	if parent != nil {
		parentPtr = parent.UnsafePointer()
	}
	w.surface = glsurface.New(parentPtr, glsurface.Handlers{
		Init:    w.initGL,
		Paint:   w.paintGL,
		Resize:  w.resizeGL,
		Show:    w.showHandler,
		Move:    w.scheduleRefresh,
		Press:   w.mousePress,
		Drag:    w.mouseDrag,
		Release: w.mouseRelease,
	})
	w.QWidget = qt.UnsafeNewQWidget(w.surface.Pointer())

	w.refreshTimer = qt.NewQTimer2(w.QObject)
	w.refreshTimer.SetSingleShot(true)
	w.refreshTimer.OnTimeout(func() {
		w.refreshScheduled = false
		w.doCapture()
	})

	registerCard(w)
	return w
}

// showHandler re-captures the backdrop when the card becomes visible, unless an
// in-progress capture just re-showed it (see captureVisibilityGuard).
func (w *AcrylicOpenGLWidget) showHandler() {
	if captureVisibilityGuarded() {
		return
	}
	w.scheduleRefresh()
}

// mousePress starts a drag if the card is draggable.
func (w *AcrylicOpenGLWidget) mousePress(button, globalX, globalY int) {
	if !w.draggable || button != int(qt.LeftButton) {
		return
	}
	w.dragging = true
	w.dragCache = nil
	w.dragPressGlobalX = globalX
	w.dragPressGlobalY = globalY
	pos := w.Pos()
	w.dragPressPosX = pos.X()
	w.dragPressPosY = pos.Y()
	w.Raise()
	moveCardToTop(w)
	// One-off z-order invalidation; see acrylic.go / refreshCardsBelow.
	refreshCardsBelow(w)
}

// mouseDrag moves the card while dragging it.
func (w *AcrylicOpenGLWidget) mouseDrag(globalX, globalY int) {
	if !w.dragging {
		return
	}
	nx := w.dragPressPosX + (globalX - w.dragPressGlobalX)
	ny := w.dragPressPosY + (globalY - w.dragPressGlobalY)
	w.Move(nx, ny)
}

// mouseRelease ends a drag and re-captures the backdrop.
func (w *AcrylicOpenGLWidget) mouseRelease() {
	if !w.dragging {
		return
	}
	w.dragging = false
	w.dragCache = nil
	w.scheduleRefresh()
}

// Delete releases the underlying Qt objects.
func (w *AcrylicOpenGLWidget) Delete() {
	unregisterCard(w)
	w.refreshTimer.Stop()
	w.surface.Delete()
}

// Refresh forces a re-capture of the backdrop on the next event-loop tick.
//
// Calling it from inside an OnUpdated callback re-captures *without* firing this
// card's callback again, so the usual wiring (every card refreshing the others)
// cannot loop.
func (w *AcrylicOpenGLWidget) Refresh() {
	if insideNotify() {
		w.silentRefresh = true
	}
	w.scheduleRefresh()
}

// BackdropImage returns the last captured backdrop (mainly for tests).
func (w *AcrylicOpenGLWidget) BackdropImage() *qt.QImage { return w.sharp }

// GrabFrameBuffer returns the card's rendered pixels as RGBA8888, along with the
// width, height and row stride in bytes.
func (w *AcrylicOpenGLWidget) GrabFrameBuffer() (pixels []byte, width, height, stride int) {
	return w.surface.GrabFrameBuffer()
}

// SetTintColor sets the overlay tint colour, e.g. "#3c8cff59" (blue at 35%);
// the alpha doubles as the tint opacity.
func (w *AcrylicOpenGLWidget) SetTintColor(hex string) {
	if r, g, b, a, ok := parseHexColor(hex); ok {
		w.tintR, w.tintG, w.tintB = r, g, b
		w.tintOpacity = float64(a) / 255.0
		w.repaint()
	}
}

// SetTintRGB sets the overlay tint colour numerically (0..255 per channel).
func (w *AcrylicOpenGLWidget) SetTintRGB(r, g, b int) {
	w.tintR, w.tintG, w.tintB = clamp255(r), clamp255(g), clamp255(b)
	w.repaint()
}

// SetTintOpacity sets the tint strength (0 = none, 1 = fully opaque tint).
func (w *AcrylicOpenGLWidget) SetTintOpacity(opacity float64) {
	w.tintOpacity = opacity
	w.repaint()
}

// SetCornerRadius rounds the card's corners (radius in pixels, 0 = square).
func (w *AcrylicOpenGLWidget) SetCornerRadius(radius int) {
	if radius < 0 {
		radius = 0
	}
	w.cornerRadius = radius
	w.repaint()
}

// SetBorderColor sets the border colour from a hex string, e.g. "#ffffff80".
func (w *AcrylicOpenGLWidget) SetBorderColor(hex string) {
	if r, g, b, a, ok := parseHexColor(hex); ok {
		w.borderR, w.borderG, w.borderB, w.borderA = r, g, b, a
		w.hasBorder = true
		w.repaint()
	}
}

// SetBorderRGB sets the border colour numerically (0..255 per channel).
func (w *AcrylicOpenGLWidget) SetBorderRGB(r, g, b, a int) {
	w.borderR, w.borderG, w.borderB, w.borderA = clamp255(r), clamp255(g), clamp255(b), clamp255(a)
	w.hasBorder = true
	w.repaint()
}

// SetBorderWidth sets the border width in pixels (0 disables the border).
func (w *AcrylicOpenGLWidget) SetBorderWidth(width int) {
	if width < 0 {
		width = 0
	}
	w.borderWidth = width
	w.repaint()
}

// SetBlurRadius sets the shader kernel radius (clamped to 0..8).
func (w *AcrylicOpenGLWidget) SetBlurRadius(radius int) {
	if radius < 0 {
		radius = 0
	}
	if radius > 8 {
		radius = 8
	}
	w.blurRadius = radius
	w.repaint()
}

// SetDownsampleFactor sets the downscale factor applied before uploading.
func (w *AcrylicOpenGLWidget) SetDownsampleFactor(factor int) {
	if factor < 1 {
		factor = 1
	}
	w.downsampleFactor = factor
	w.Refresh()
}

// SetNoiseAmount sets the shader grain strength (0 disables it).
func (w *AcrylicOpenGLWidget) SetNoiseAmount(amount int) {
	if amount < 0 {
		amount = 0
	}
	w.noiseAmount = amount
	w.repaint()
}

// SetDraggable enables mouse dragging of the widget; the backdrop follows the
// card live while it is dragged.
func (w *AcrylicOpenGLWidget) SetDraggable(draggable bool) {
	w.draggable = draggable
}

// SetCaption sets the label shown in the top-left corner.
//
// It is a real QLabel child of the card, not text drawn into the GL surface. Qt's
// GL paint engine rasterises glyphs into a texture that is then scaled by the
// device pixel ratio, so a caption painted there looks bold and blurry next to the
// CPU back end (and next to the crisp text Qt draws for ordinary widgets) — very
// visible on a 150% display. A label goes through the normal widget pipeline and
// is crisp at any DPI, and it can be styled with a stylesheet like any widget.
func (w *AcrylicOpenGLWidget) SetCaption(text string) {
	w.caption = text

	if w.captionLabel == nil {
		label := qt.NewQLabel3(text)
		label.SetParent(w.QWidget)
		label.SetStyleSheet("QLabel { color: #ffffff; background: transparent; }")

		font := captionFont(w.QWidget)
		label.SetFont(font)
		font.Delete()

		label.AdjustSize()
		// Matches the baseline the other back ends paint their caption at.
		label.Move(16, 14)
		w.captionLabel = label
	} else {
		w.captionLabel.SetText(text)
		w.captionLabel.AdjustSize()
	}
	w.captionLabel.SetVisible(text != "")
	w.repaint()
}

func (w *AcrylicOpenGLWidget) repaint() {
	if w.surface != nil {
		w.surface.Update()
	}
}

func (w *AcrylicOpenGLWidget) scheduleRefresh() {
	if w.capturing || w.refreshScheduled || captureVisibilityGuarded() || refreshingSelf(w) {
		return
	}
	w.refreshScheduled = true
	w.refreshTimer.Start(0)
}

func (w *AcrylicOpenGLWidget) doCapture() {
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
	// Kept for sibling compositing: the shader blurs on the GPU, so without this
	// copy another card's rounded corner would reveal the unblurred backdrop.
	w.composite = downsampleAndBlur(img, w.downsampleFactor, w.blurRadius, w.noiseAmount, w.blurSigma)
	w.textureDirty = true
	w.repaint()

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
//
// If several cards in different windows should follow the same source, this is the
// place to call acrylic.RefreshWindow(otherWindow) or acrylic.RefreshAll().
func (w *AcrylicOpenGLWidget) OnUpdated(fn func()) {
	w.onUpdated = fn
}

func (w *AcrylicOpenGLWidget) notifyUpdated() {
	notifyCardUpdated(w, w.onUpdated)
}

// backdropCachePtr returns the per-drag composite cache while the card is being
// dragged, else nil.
func (w *AcrylicOpenGLWidget) backdropCachePtr() **qt.QImage {
	if w.dragging {
		return &w.dragCache
	}
	return nil
}

// widget returns the underlying QWidget (used for sibling classification).
func (w *AcrylicOpenGLWidget) widget() *qt.QWidget { return w.QWidget }

// qtPaintsCard reports that Qt paints this card like any other child widget,
// because the surface is not a native window.
func (w *AcrylicOpenGLWidget) qtPaintsCard() bool { return true }

// drawAcrylic paints this card's acrylic content at the given offset for sibling
// compositing.
//
// Unlike the other back ends this one needs no approximation: the card renders
// into its own framebuffer, so its exact pixels — shader blur, tint, border,
// rounded corners and child widgets included — can simply be composited. The
// fallback below only runs if the surface has no usable context yet.
func (w *AcrylicOpenGLWidget) drawAcrylic(painter *qt.QPainter, x, y int) {
	if img := w.framebufferImage(); img != nil {

		rect := qt.NewQRect4(x, y, w.Width(), w.Height())
		painter.DrawImage6(rect, img)
		rect.Delete()
		img.Delete()

		// grabFramebuffer returns the GL content only; Qt composites the child
		// widgets over it separately, so they are added here.
		painter.Save()
		renderChildWidgets(w.QWidget, painter, x, y)
		painter.Restore()
		return
	}

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
	renderChildWidgets(w.QWidget, painter, x, y)

	painter.Restore()

	drawShaderStyleBorder(painter, x, y, width, height,
		w.hasBorder, w.borderWidth, w.cornerRadius,
		w.borderR, w.borderG, w.borderB, w.borderA)
}

// framebufferImage renders the card into its framebuffer and returns the result
// as an RGBA8888 image owned by the caller, or nil when there is no context yet.
func (w *AcrylicOpenGLWidget) framebufferImage() *qt.QImage {
	pixels, width, height, stride := w.surface.GrabFrameBuffer()
	if width <= 0 || height <= 0 || len(pixels) < stride*height {
		return nil
	}

	img := qt.NewQImage3(width, height, qt.QImage__Format_RGBA8888)
	if n := img.ByteCount(); n > 0 {
		dst := unsafe.Slice(img.Bits(), n)
		rowBytes := width * 4
		if lb := img.BytesPerLine(); lb < rowBytes {
			rowBytes = lb
		}
		if rowBytes > stride {
			rowBytes = stride
		}
		// grabFramebuffer already returns the image top-down, matching QImage.
		for row := 0; row < height; row++ {
			src := pixels[row*stride:]
			if rowBytes > len(src) {
				rowBytes = len(src)
			}
			copy(dst[row*img.BytesPerLine():], src[:rowBytes])
		}
	}
	return img
}

func (w *AcrylicOpenGLWidget) initGL() {
	if err := gl.Init(); err != nil {
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

func (w *AcrylicOpenGLWidget) resizeGL(width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (w *AcrylicOpenGLWidget) paintGL() {
	if w.glInitialized {
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		if w.textureDirty {
			w.uploadTextures()
			w.textureDirty = false
		}
		if w.texture != 0 && w.sharpTexture != 0 {
			w.drawAcrylicQuad()
		}
	}

	// The caption is a real label child (see SetCaption), so Qt paints it with the
	// normal (crisp) text path. Child widgets need no handling here either: Qt
	// composites them normally, because this surface is not a native window.
}

// drawAcrylicQuad binds the two textures, sets every uniform and draws the
// full-size quad.
func (w *AcrylicOpenGLWidget) drawAcrylicQuad() {
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

// uploadTextures uploads the downsampled blur texture (unit 0) and the
// full-resolution sharp texture used outside the rounded corners (unit 1).
func (w *AcrylicOpenGLWidget) uploadTextures() {
	if w.cached != nil {
		flipped := w.cached.Mirrored2(false, true)
		w.texWidth = flipped.Width()
		w.texHeight = flipped.Height()
		if w.texWidth > 0 && w.texHeight > 0 {
			w.texelX = 1.0 / float32(w.texWidth)
			w.texelY = 1.0 / float32(w.texHeight)
		}

		if w.texture == 0 {
			gl.GenTextures(1, &w.texture)
		}
		gl.BindTexture(gl.TEXTURE_2D, w.texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

		if n := flipped.ByteCount(); n > 0 {
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
				int32(w.texWidth), int32(w.texHeight), 0,
				gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(flipped.Bits()))
		}
	}

	if w.sharp != nil {
		sflipped := w.sharp.Mirrored2(false, true)
		if w.sharpTexture == 0 {
			gl.GenTextures(1, &w.sharpTexture)
		}
		gl.BindTexture(gl.TEXTURE_2D, w.sharpTexture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

		if n := sflipped.ByteCount(); n > 0 {
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
				int32(sflipped.Width()), int32(sflipped.Height()), 0,
				gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(sflipped.Bits()))
		}
	}

	gl.ActiveTexture(gl.TEXTURE0)

	// Restore the packing state we changed: Qt's own painter uploads (the glyph
	// atlas for the caption, for instance) go through the same context, and a
	// leaked UNPACK_ALIGNMENT of 1 makes the glyph rows line up differently — the
	// caption came out noticeably bolder than the CPU back end's.
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

