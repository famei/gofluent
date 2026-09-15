// Package acrylic provides "Acrylic" (frosted-glass) background widgets for Qt
// applications built with the miqt Go bindings.
//
// Two widgets are provided:
//
//   - AcrylicWidget: a plain QWidget that captures the pixels behind it, blurs
//     them on the CPU with a separable Gaussian, tints them and paints the
//     result. It has no OpenGL dependency.
//   - AcrylicGLWidget: a QGLWidget that uploads the captured pixels as a
//     texture and performs the blur on the GPU with a GLSL fragment shader
//     (see shaders.go).
//
// AcrylicWidget captures its backdrop by rendering the top-level window
// (children included) while suppressing its own paint — deterministic, no
// flicker, and overlapping cards blend through each other.
//
// AcrylicGLWidget is a native QGLWidget, so it cannot be rendered this way; it
// briefly hides itself and grabs the desktop region behind its geometry instead.
package acrylic

import (
	"sync"
	"unsafe"

	"github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/acrylic/internal/blur"
)

// acrylicCard is the common interface of the CPU and GPU acrylic widgets used
// for sibling compositing.
type acrylicCard interface {
	Width() int
	Height() int
	Window() *qt.QWidget
	MapTo(*qt.QWidget, *qt.QPoint) *qt.QPoint
	drawAcrylic(*qt.QPainter, int, int)
	// widget returns the underlying QWidget, used to tell acrylic cards apart
	// from ordinary child widgets while compositing.
	widget() *qt.QWidget
	// qtPaintsCard reports whether Qt paints this card as ordinary widget
	// content. It is true for the non-native back ends, which therefore have to
	// be hidden while another card captures its backdrop; a native QGLWidget is
	// left unpainted inside a parent render anyway.
	qtPaintsCard() bool
	// Refresh asks the card to re-capture its backdrop; used to keep the cards
	// stacked above a changed card in sync with it.
	Refresh()
}

// acrylicWidgetSet returns the raw pointers of every registered acrylic card's
// QWidget, so ordinary children can be told apart from cards.
func acrylicWidgetSet() map[unsafe.Pointer]bool {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	set := make(map[unsafe.Pointer]bool, len(cardRegistry.cards))
	for _, c := range cardRegistry.cards {
		if w := c.widget(); w != nil {
			set[w.UnsafePointer()] = true
		}
	}
	return set
}

// captureBackdropFor builds the backdrop behind card c: the top-level window's
// own background plus everything drawn below c — ordinary sibling widgets
// (labels, buttons, images, custom widgets…) and other acrylic cards.
//
// The window is rendered with DrawWindowBackground only, deliberately without
// DrawChildren: render() would recurse into native QGLWidget children and call
// their GL paintEvent, which corrupts their surface ("swapBuffers() called with
// non-exposed window"). Children are therefore composited explicitly:
//
//	ordinary widget -> its own grab, drawn at its position;
//	acrylic card    -> its acrylic content, drawn at its position.
//
// Cards must be created after the widgets they should blur (Qt stacks later
// children on top), which is the usual layout anyway.
func captureBackdropFor(c acrylicCard, cache **qt.QImage) *qt.QImage {
	top := c.Window()
	if top == nil {
		return nil
	}
	tw := top.Width()
	th := top.Height()
	if tw <= 0 || th <= 0 {
		return nil
	}

	// `full` is the whole window composite. While a card is being dragged the
	// background cannot change, so it is reused and each move only re-crops —
	// which keeps the acrylic content glued to the card instead of lagging
	// behind it (no drag smear).
	var full *qt.QImage
	if cache != nil && *cache != nil {
		full = *cache
	} else {
		// buildWindowComposite already attached a finalizer, so the composite is
		// released automatically once neither the cache nor this call holds it.
		full = buildWindowComposite(top, c)
		if full == nil {
			return nil
		}
		if cache != nil {
			*cache = full
		}
	}

	local := qt.NewQPoint2(0, 0)
	defer local.Delete()
	pos := c.MapTo(top, local)
	return full.Copy2(pos.X(), pos.Y(), c.Width(), c.Height())
}

// buildWindowComposite renders what is behind card c across the whole window:
// the window's own background, then every widget below c.
func buildWindowComposite(top *qt.QWidget, c acrylicCard) *qt.QImage {
	tw := top.Width()
	th := top.Height()

	img := qt.NewQImage3(tw, th, qt.QImage__Format_RGBA8888)
	img.GoGC()

	offset := qt.NewQPoint2(0, 0)
	defer offset.Delete()
	region := qt.NewQRegion2(0, 0, tw, th)
	defer region.Delete()

	// 1. The window's own background (no children).
	top.Render4(img.QPaintDevice, offset, region, qt.QWidget__DrawWindowBackground)

	painter := qt.NewQPainter2(img.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHint2(qt.QPainter__SmoothPixmapTransform, true)

	// 2. Ordinary child widgets, each painted straight onto the composite.
	//
	// The widget is rendered itself instead of being Grab()ed into a pixmap and
	// blitted: grab() renders with DrawWindowBackground, so everything the widget
	// leaves unpainted — the wedges outside a QPushButton's rounded corners, for
	// example — was filled with the widget's palette colour (white) and then
	// pasted over the backdrop, which broke the widget's rounded border into a
	// solid block. That is invisible inside the blurred acrylic but clearly
	// visible in the sharp pixels outside a card's rounded corners. Painting onto
	// the composite leaves those pixels as the backdrop that is already there.
	// DrawChildren (not DrawWindowBackground) is what makes that work; the whole
	// window background was laid down in step 1.
	//
	// Cards are usually nested inside a container (a QStackedWidget page, a layout
	// panel…), and the container is a direct child of the window: rendering it
	// would paint the card into its own backdrop. Each capture then contained the
	// previous frame, so the acrylic drifted towards a flat tint and the rounded
	// corners sampled the card itself instead of the content behind it.
	//
	// The cards are therefore masked away for the duration of this render. Cutting
	// the cards' rectangles out of the render region instead would also cut out
	// whatever lies *under* them — the background widgets the acrylic is supposed
	// to blur — so the corner pixels ended up showing the bare window background.
	// Native GL cards are skipped: Qt already leaves their area unpainted inside a
	// parent render.
	acrylics := acrylicWidgetSet()
	restoreCards := parkCardsForCapture(top, c)
	restoreEffects := suspendGraphicsEffects(top)

	// While the cards are parked outside the window nothing opaque covers the area
	// a card occupies, so the container paints its background widgets there like
	// anywhere else and the backdrop comes out complete. (An earlier version parked
	// nothing and instead rendered the children a second time, clipped to the card,
	// to fill the hole Qt leaves for an opaque child — that second pass painted the
	// content shifted inside the card's rectangle.)
	renderWindowChildren(top, painter, acrylics)

	// The cards are back in place before step 3: the GPU back end reads its own
	// framebuffer there, which needs the surface alive.
	restoreCards()
	restoreEffects()

	// 3. Acrylic cards stacked beneath this one (bottom -> top).
	for _, sib := range siblingsBelow(c) {
		local := qt.NewQPoint2(0, 0)
		pos := sib.MapTo(top, local)
		local.Delete()
		sib.drawAcrylic(painter, pos.X(), pos.Y())
	}

	painter.End()
	return img
}

// renderWindowChildren paints every direct child of the window onto painter.
func renderWindowChildren(top *qt.QWidget, painter *qt.QPainter, acrylics map[unsafe.Pointer]bool) {
	for _, obj := range top.Children() {
		if obj == nil || !obj.IsWidgetType() {
			continue
		}
		child := qt.UnsafeNewQWidget(obj.UnsafePointer())
		if child == nil || child.IsWindow() || !child.IsVisible() {
			continue
		}
		if acrylics[child.UnsafePointer()] {
			continue // acrylic cards are composited in step 3
		}
		local := qt.NewQPoint2(0, 0)
		pos := child.MapTo(top, local)
		local.Delete()
		region := qt.NewQRegion2(0, 0, child.Width(), child.Height())

		if hasGraphicsEffect(child, top) {
			// The effects are disabled for the duration of the capture, but Qt still
			// routes an effect-bearing subtree through its offscreen pixmap path in
			// some situations, where render()'s target offset is lost. Translating the
			// painter instead keeps the content in place.
			painter.Save()
			painter.Translate2(float64(pos.X()), float64(pos.Y()))
			origin := qt.NewQPoint2(0, 0)
			child.Render7(painter, origin, region, qt.QWidget__DrawChildren)
			origin.Delete()
			painter.Restore()
		} else {
			child.Render7(painter, pos, region, qt.QWidget__DrawChildren)
		}
		region.Delete()
	}
}

// suspendGraphicsEffects disables every enabled QGraphicsEffect below the given
// widget for the duration of a capture, and returns a function that enables them
// again.
//
// Qt renders an effect-bearing subtree into an offscreen pixmap, and while doing
// so it ignores render()'s target offset (and masks): the captured content landed
// shifted by the container's position. With the effects disabled the subtree
// renders through the ordinary path and everything lines up. Disabling does not
// delete the effect, and no frame is presented before it is re-enabled, so nothing
// flickers.
func suspendGraphicsEffects(top *qt.QWidget) func() {
	var effects []*qt.QGraphicsEffect

	var walk func(w *qt.QWidget)
	walk = func(w *qt.QWidget) {
		if w == nil {
			return
		}
		if e := w.GraphicsEffect(); e != nil && e.IsEnabled() {
			e.SetEnabled(false)
			effects = append(effects, e)
		}
		for _, obj := range w.Children() {
			if obj == nil || !obj.IsWidgetType() {
				continue
			}
			walk(qt.UnsafeNewQWidget(obj.UnsafePointer()))
		}
	}
	walk(top)

	if len(effects) == 0 {
		return func() {}
	}

	restored := false
	return func() {
		if restored {
			return
		}
		restored = true
		for _, e := range effects {
			e.SetEnabled(true)
		}
	}
}

// hasGraphicsEffect reports whether the widget or any of its ancestors below the
// given window carries a QGraphicsEffect. Those subtrees are rendered by Qt
// through an offscreen pixmap.
func hasGraphicsEffect(w *qt.QWidget, stopAt *qt.QWidget) bool {
	stop := stopAt.UnsafePointer()
	for cur := w; cur != nil; cur = cur.ParentWidget() {
		if cur.GraphicsEffect() != nil {
			return true
		}
		if cur.UnsafePointer() == stop {
			break
		}
	}
	return false
}

// registeredCards returns a snapshot of the card registry, in z order.
func registeredCards() []acrylicCard {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	return append([]acrylicCard(nil), cardRegistry.cards...)
}

// RefreshAll re-captures every acrylic card. Call it when the content behind cards
// in more than one window changed at once: a card re-capturing on its own already
// refreshes the cards stacked above it in the same window, but nothing crosses
// window boundaries.
func RefreshAll() {
	for _, c := range registeredCards() {
		c.Refresh()
	}
}

// RefreshWindow re-captures every acrylic card that lives in the same top-level
// window as the given widget. Pass the card's window — or any widget inside it,
// such as the container that holds the cards: the argument is resolved to its
// top-level window first.
func RefreshWindow(w *qt.QWidget) {
	if w == nil {
		return
	}
	top := w.Window()
	if top == nil {
		return
	}
	for _, c := range cardsInWindow(top) {
		c.Refresh()
	}
}

// notifyingCard is the card whose OnUpdated callback is currently running. A card
// that refreshes itself from inside its own callback would loop forever (capture →
// callback → Refresh → capture …), so its own Refresh is ignored while its callback
// runs: it has just captured, so there is nothing to redo.
var notifyingCard acrylicCard

// notifyDepth counts the OnUpdated callbacks currently on the stack.
var notifyDepth = struct {
	sync.Mutex
	depth int
}{}

// notifyCardUpdated runs a card's OnUpdated callback with the self-refresh guard
// installed.
func notifyCardUpdated(c acrylicCard, fn func()) {
	if fn == nil {
		return
	}
	previous := notifyingCard
	notifyingCard = c
	notifyDepth.Lock()
	notifyDepth.depth++
	notifyDepth.Unlock()

	defer func() {
		notifyingCard = previous
		notifyDepth.Lock()
		notifyDepth.depth--
		notifyDepth.Unlock()
	}()

	fn()
}

// refreshingSelf reports whether c is the card whose OnUpdated callback is running.
func refreshingSelf(c acrylicCard) bool {
	return notifyingCard == c
}

// insideNotify reports whether we are inside an OnUpdated callback. A refresh asked
// for from there is a *follow-up* refresh: the cards it re-captures do not fire
// their own callbacks again.
//
// Without that rule the obvious wiring — every card refreshing all the other cards
// when it updates — never stops: capture → callback → refresh the others → their
// callbacks → refresh the first one → …, an endless ring. The automatic chain
// (cards above a re-captured card) is not affected: it is issued outside a callback
// and keeps propagating.
func insideNotify() bool {
	notifyDepth.Lock()
	defer notifyDepth.Unlock()
	return notifyDepth.depth > 0
}

// cardsInWindow returns every registered acrylic card that lives in the given
// window, in z order.
func cardsInWindow(top *qt.QWidget) []acrylicCard {
	if top == nil {
		return nil
	}
	topPtr := top.UnsafePointer()

	cardRegistry.Lock()
	cards := append([]acrylicCard(nil), cardRegistry.cards...)
	cardRegistry.Unlock()

	var out []acrylicCard
	for _, c := range cards {
		if w := c.Window(); w != nil && w.UnsafePointer() == topPtr {
			out = append(out, c)
		}
	}
	return out
}

// captureVisibilityGuard counts the cards currently hidden by an in-progress
// capture. Showing a card again fires its showEvent, which would schedule
// another refresh — and that refresh captures again, hides the card again, and
// so on: an endless capture loop that pegs a CPU core. While the guard is up a
// card therefore skips the refresh its showEvent asks for; it has just been
// captured anyway, so nothing is lost.
var captureVisibilityGuard = struct {
	sync.Mutex
	count int
}{}

func beginCaptureVisibilityGuard() {
	captureVisibilityGuard.Lock()
	captureVisibilityGuard.count++
	captureVisibilityGuard.Unlock()
}

func endCaptureVisibilityGuard() {
	captureVisibilityGuard.Lock()
	captureVisibilityGuard.count--
	captureVisibilityGuard.Unlock()
}

func captureVisibilityGuarded() bool {
	captureVisibilityGuard.Lock()
	defer captureVisibilityGuard.Unlock()
	return captureVisibilityGuard.count > 0
}

// parkCardsForCapture moves the card being captured — and every card stacked above
// it — out of the window for the duration of a capture, and returns a function
// that puts them back.
//
// Only those cards matter: whatever lies below the card belongs in its backdrop, so
// cards below it may stay where they are. The card itself and the cards above it
// must not appear in its own backdrop.
//
// Parking — rather than hiding or masking — is what keeps dragging alive: hiding a
// widget drops its mouse grab (and the card being dragged is one of them), while a
// mask is ignored by QWidget::render(), so a masked card is still painted into its
// own backdrop. Moving a widget touches neither visibility, focus nor the grab, and
// it is undone before any frame is presented. The refreshes a move would trigger
// are suppressed by the capture guard, which would otherwise capture again and loop.
func parkCardsForCapture(top *qt.QWidget, capturing acrylicCard) func() {
	type parkedCard struct {
		widget *qt.QWidget
		x, y   int
	}

	topPtr := top.UnsafePointer()
	reachedCapturing := false

	var parked []parkedCard
	for _, c := range registeredCards() {
		if w := c.Window(); w == nil || w.UnsafePointer() != topPtr {
			continue
		}

		// A native GL card (QGLWidget) inside a container must always be parked:
		// rendering that container paints the native surface at the wrong time
		// ("swapBuffers() called with non-exposed window") and degrades the card —
		// and its pixels cannot be captured through a QPainter anyway.
		nativeInContainer := false
		if !c.qtPaintsCard() {
			if w := c.widget(); w != nil {
				parent := w.ParentWidget()
				nativeInContainer = parent != nil && parent.UnsafePointer() != topPtr
			}
		}

		if c == capturing {
			reachedCapturing = true
		} else if !reachedCapturing && !nativeInContainer {
			continue // stacked below: its content belongs in the backdrop
		}

		w := c.widget()
		if w == nil || !w.IsVisible() {
			continue
		}
		// A native card that is a direct child of the window is left alone: Qt never
		// renders native children there, and hiding or moving it would be pointless.
		if !c.qtPaintsCard() && !nativeInContainer {
			continue
		}
		pos := w.Pos()
		parked = append(parked, parkedCard{widget: w, x: pos.X(), y: pos.Y()})
	}
	if len(parked) == 0 {
		return func() {}
	}

	const away = -20000

	beginCaptureVisibilityGuard()
	for _, p := range parked {
		p.widget.Move(away, away)
	}

	restored := false
	return func() {
		if restored {
			return
		}
		restored = true
		for _, p := range parked {
			p.widget.Move(p.x, p.y)
		}
		endCaptureVisibilityGuard()
	}
}

// cardRegistry tracks acrylic cards in z order (bottom -> top) so a card can
// composite the cards beneath it when capturing its backdrop.
var cardRegistry = struct {
	sync.Mutex
	cards []acrylicCard
}{}

func registerCard(c acrylicCard) {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	cardRegistry.cards = append(cardRegistry.cards, c)
}

func unregisterCard(c acrylicCard) {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	for i, x := range cardRegistry.cards {
		if x == c {
			cardRegistry.cards = append(cardRegistry.cards[:i], cardRegistry.cards[i+1:]...)
			break
		}
	}
}

// moveCardToTop moves c to the end (top) of the registry to mirror Qt's z order
// after the card is raised.
func moveCardToTop(c acrylicCard) {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	for i, x := range cardRegistry.cards {
		if x == c {
			cardRegistry.cards = append(cardRegistry.cards[:i], cardRegistry.cards[i+1:]...)
			break
		}
	}
	cardRegistry.cards = append(cardRegistry.cards, c)
}

// renderChildWidgets paints a widget's direct child widgets onto painter, offset
// by (x, y). A card's children sit on top of its acrylic, so they belong to the
// picture another card sees when it composites this one.
func renderChildWidgets(parent *qt.QWidget, painter *qt.QPainter, x, y int) {
	if parent == nil {
		return
	}
	for _, obj := range parent.Children() {
		if obj == nil || !obj.IsWidgetType() {
			continue
		}
		child := qt.UnsafeNewQWidget(obj.UnsafePointer())
		if child == nil || child.IsWindow() || !child.IsVisible() {
			continue
		}
		zero := qt.NewQPoint2(0, 0)
		local := child.MapTo(parent, zero)
		zero.Delete()

		region := qt.NewQRegion2(0, 0, child.Width(), child.Height())
		at := qt.NewQPoint2(x+local.X(), y+local.Y())
		child.Render7(painter, at, region, qt.QWidget__DrawChildren)
		at.Delete()
		region.Delete()
	}
}

// drawShaderStyleBorder draws a border the way the GLSL shader does: a band
// hugging the inside of the shape whose outermost pixel is only half covered
// (the shader's anti-aliasing band) and whose inner edge is hard.
//
// A stroked QPen would instead sit half a pixel wider and bolder than the card
// itself looks, so a sibling card's rounded corner — which samples these pixels
// — would show the border thickened where the two met.
func drawShaderStyleBorder(painter *qt.QPainter, x, y, width, height int,
	hasBorder bool, borderWidth, cornerRadius, r, g, b, a int) {
	if !hasBorder || borderWidth <= 0 || width <= 0 || height <= 0 {
		return
	}

	bw := float64(borderWidth)
	rOuter := float64(cornerRadius) - 0.5
	rInner := float64(cornerRadius) - bw
	if rOuter < 0 {
		rOuter = 0
	}
	if rInner < 0 {
		rInner = 0
	}

	outer := qt.NewQPainterPath()
	outer.AddRoundedRect2(
		float64(x)+0.5, float64(y)+0.5,
		float64(width)-1, float64(height)-1, rOuter, rOuter)
	inner := qt.NewQPainterPath()
	inner.AddRoundedRect2(
		float64(x)+bw, float64(y)+bw,
		float64(width)-2*bw, float64(height)-2*bw, rInner, rInner)

	ring := outer.Subtracted(inner)
	colour := qt.NewQColor11(r, g, b, a)
	brush := qt.NewQBrush3(colour)
	painter.SetRenderHint2(qt.QPainter__Antialiasing, true)
	painter.FillPath(ring, brush)

	brush.Delete()
	colour.Delete()
	inner.Delete()
	outer.Delete()
}

// cardsAbove returns the acrylic cards in the same window stacked above c (in
// bottom -> top order). Together with siblingsBelow it splits the registry
// around c.
func cardsAbove(c acrylicCard) []acrylicCard {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	top := c.Window()
	if top == nil {
		return nil
	}
	topPtr := top.UnsafePointer()
	var out []acrylicCard
	above := false
	for _, x := range cardRegistry.cards {
		if x == c {
			above = true
			continue
		}
		if !above {
			continue
		}
		if xw := x.Window(); xw != nil && xw.UnsafePointer() == topPtr {
			out = append(out, x)
		}
	}
	return out
}

// refreshCardsAbove re-captures every acrylic card stacked above c. c's rendered
// acrylic is part of their backdrop, so once c changes (moved, resized or
// re-tinted) they would keep showing a stale copy of it — most visibly in the
// sharp pixels outside their rounded corners. The walk only ever goes upward, so
// it always terminates: a card never schedules itself again through this path.
func refreshCardsAbove(c acrylicCard) {
	for _, x := range cardsAbove(c) {
		x.Refresh()
	}
}

// refreshCardsBelow re-captures every acrylic card stacked below c.
//
// It is what a z-order change needs: while c sits below them, the cards above have
// c inside their backdrop, and the moment c is raised they must drop it again. That
// happens once per raise (at the start of a drag), which is what keeps dragging
// cheap — refreshing every card on every drag step would mean one full window
// composite per mouse move, per card.
//
// A plain repaint request (Update) is *not* enough here, and that was measured: the
// card has just been covered by the raised one, so Qt has no exposed area to paint
// and silently skips the paint event — the sibling keeps showing the stale acrylic
// on screen until something else forces it (verify18: screen pixel stayed red after
// the press + move, and turned green after a forced Refresh()).
func refreshCardsBelow(c acrylicCard) {
	for _, x := range siblingsBelow(c) {
		x.Refresh()
	}
}

// siblingsBelow returns the acrylic cards in the same window stacked beneath c
// (in bottom -> top order).
func siblingsBelow(c acrylicCard) []acrylicCard {
	cardRegistry.Lock()
	defer cardRegistry.Unlock()
	top := c.Window()
	if top == nil {
		return nil
	}
	topPtr := top.UnsafePointer()
	var out []acrylicCard
	for _, x := range cardRegistry.cards {
		if x == c {
			break
		}
		xw := x.Window()
		if xw != nil && xw.UnsafePointer() == topPtr {
			out = append(out, x)
		}
	}
	return out
}

// downsampleAndBlur scales img down by factor, blurs the small copy in place,
// adds grain, and returns it. The caller draws it scaled back up with a smooth
// transform.
func downsampleAndBlur(img *qt.QImage, factor, radius, noiseAmount int, sigma float64) *qt.QImage {
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

	if n := small.ByteCount(); n > 0 {
		buf := unsafe.Slice(small.Bits(), n)
		blur.BlurRGBA(buf, small.BytesPerLine(), small.Width(), small.Height(), radius, sigma)
		blur.AddNoise(buf, small.BytesPerLine(), small.Width(), small.Height(), noiseAmount)
	}
	return small
}

// AcrylicWidget is a QWidget that paints a blurred, tinted copy of whatever is
// currently behind it.
type AcrylicWidget struct {
	*qt.QWidget

	cached *qt.QImage // downscaled + blurred backdrop, painted at full size

	caption string

	downsampleFactor    int
	blurRadius          int
	blurSigma           float64
	noiseAmount         int
	tintR, tintG, tintB int
	tintOpacity         float64

	cornerRadius                       int
	borderWidth                        int
	hasBorder                          bool
	borderR, borderG, borderB, borderA int

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

// NewAcrylicWidget creates an AcrylicWidget with sensible defaults. The widget
// captures and paints its backdrop automatically when shown, moved or resized;
// call Refresh() whenever the content behind it changes.
func NewAcrylicWidget(parent *qt.QWidget) *AcrylicWidget {
	w := &AcrylicWidget{
		QWidget:          qt.NewQWidget(parent),
		downsampleFactor: 3,
		blurRadius:       5,
		blurSigma:        2.0,
		noiseAmount:      2,
		tintR:            255,
		tintG:            255,
		tintB:            255,
		tintOpacity:      0.30,
	}

	w.refreshTimer = qt.NewQTimer2(w.QObject)
	w.refreshTimer.SetSingleShot(true)
	w.refreshTimer.OnTimeout(func() {
		w.refreshScheduled = false
		w.doCapture()
	})

	w.OnPaintEvent(func(super func(ev *qt.QPaintEvent), ev *qt.QPaintEvent) {
		w.paint()
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
		// The CPU capture is an off-screen render (no hide/grab/show), so the
		// backdrop follows the card live while it is dragged.
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
		// One-off: the cards that had this one underneath must drop it from their
		// backdrops. Doing it here (once per raise) instead of on every drag step
		// is what keeps dragging cheap.
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
func (w *AcrylicWidget) Delete() {
	unregisterCard(w)
	w.refreshTimer.Stop()
	w.QWidget.Delete()
}

// Refresh forces a re-capture of the backdrop on the next event-loop tick.
//
// Calling it from inside an OnUpdated callback re-captures *without* firing this
// card's callback again, so the usual wiring (every card refreshing the others)
// cannot loop.
func (w *AcrylicWidget) Refresh() {
	if insideNotify() {
		w.silentRefresh = true
	}
	w.scheduleRefresh()
}

// SetTintColor sets the overlay tint from a hex string, e.g. "#60a0ff",
// "#60a0ff80", "#f0a" or "#f0a8". When the string carries an alpha component it
// also sets the tint opacity, so a single call is enough:
//
//	panel.SetTintColor("#60a0ff80") // blue, 50% tint
func (w *AcrylicWidget) SetTintColor(hex string) {
	if r, g, b, a, ok := parseHexColor(hex); ok {
		w.tintR, w.tintG, w.tintB = r, g, b
		w.tintOpacity = float64(a) / 255.0
		w.Update()
	}
}

// SetTintRGB sets the overlay tint colour numerically (0..255 per channel).
func (w *AcrylicWidget) SetTintRGB(r, g, b int) {
	w.tintR, w.tintG, w.tintB = clamp255(r), clamp255(g), clamp255(b)
	w.Update()
}

// SetTintOpacity sets the tint strength (0 = none, 1 = fully opaque tint).
func (w *AcrylicWidget) SetTintOpacity(opacity float64) {
	w.tintOpacity = opacity
	w.Update()
}

// SetCornerRadius rounds the card's corners (radius in pixels, 0 = square).
func (w *AcrylicWidget) SetCornerRadius(radius int) {
	if radius < 0 {
		radius = 0
	}
	w.cornerRadius = radius
	w.Update()
}

// SetBorderColor sets the border colour from a hex string, e.g. "#ffffff80".
func (w *AcrylicWidget) SetBorderColor(hex string) {
	if r, g, b, a, ok := parseHexColor(hex); ok {
		w.borderR, w.borderG, w.borderB, w.borderA = r, g, b, a
		w.hasBorder = true
		w.Update()
	}
}

// SetBorderRGB sets the border colour numerically (0..255 per channel).
func (w *AcrylicWidget) SetBorderRGB(r, g, b, a int) {
	w.borderR, w.borderG, w.borderB, w.borderA = clamp255(r), clamp255(g), clamp255(b), clamp255(a)
	w.hasBorder = true
	w.Update()
}

// SetBorderWidth sets the border width in pixels (0 disables the border).
func (w *AcrylicWidget) SetBorderWidth(width int) {
	if width < 0 {
		width = 0
	}
	w.borderWidth = width
	w.Update()
}

// SetBlurRadius sets the Gaussian radius applied on the downscaled image.
func (w *AcrylicWidget) SetBlurRadius(radius int) {
	w.blurRadius = radius
	w.Refresh()
}

// SetDownsampleFactor sets the downscale factor used before blurring.
func (w *AcrylicWidget) SetDownsampleFactor(factor int) {
	if factor < 1 {
		factor = 1
	}
	w.downsampleFactor = factor
	w.Refresh()
}

// SetNoiseAmount sets the grain strength applied after blurring (0 disables).
func (w *AcrylicWidget) SetNoiseAmount(amount int) {
	if amount < 0 {
		amount = 0
	}
	w.noiseAmount = amount
	w.Refresh()
}

// SetDraggable enables mouse dragging of the widget. The backdrop follows the
// card live while it is dragged.
func (w *AcrylicWidget) SetDraggable(draggable bool) {
	w.draggable = draggable
}

// SetCaption sets an optional label drawn in the top-left corner of the widget.
func (w *AcrylicWidget) SetCaption(text string) {
	w.caption = text
	w.Update()
}

// BackdropImage returns the current blurred backdrop image (nil before the
// first capture completes). Primarily useful for testing/saving the result.
func (w *AcrylicWidget) BackdropImage() *qt.QImage {
	return w.cached
}

func (w *AcrylicWidget) scheduleRefresh() {
	if w.capturing || w.refreshScheduled || captureVisibilityGuarded() || refreshingSelf(w) {
		return
	}
	w.refreshScheduled = true
	w.refreshTimer.Start(0)
}

func (w *AcrylicWidget) doCapture() {
	if w.capturing {
		return
	}
	w.capturing = true
	defer func() { w.capturing = false }()

	img := captureBackdropFor(w, w.backdropCachePtr())
	if img == nil {
		return
	}
	w.cached = downsampleAndBlur(img, w.downsampleFactor, w.blurRadius, w.noiseAmount, w.blurSigma)
	w.Update()

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
// It is the hook for keeping *other* acrylic widgets in sync. Cards stacked above
// this one in the same window are refreshed automatically; this callback is for
// everything the library cannot know about — cards in another window, a custom
// widget drawing the same backdrop, a status display, and so on. The callback runs
// on the GUI thread right after the card is up to date, and it may fire often
// (once per capture, so once per drag step), so keep it cheap: call Refresh() on
// the other widget rather than doing work there.
//
// Pass nil to remove a previously registered callback.
func (w *AcrylicWidget) OnUpdated(fn func()) {
	w.onUpdated = fn
}

func (w *AcrylicWidget) notifyUpdated() {
	notifyCardUpdated(w, w.onUpdated)
}

// backdropCachePtr returns the per-drag composite cache while the card is being
// dragged, else nil.
func (w *AcrylicWidget) backdropCachePtr() **qt.QImage {
	if w.dragging {
		return &w.dragCache
	}
	return nil
}

// widget returns the underlying QWidget (used for sibling classification).
func (w *AcrylicWidget) widget() *qt.QWidget { return w.QWidget }

// qtPaintsCard reports that Qt paints this card like any other child widget.
func (w *AcrylicWidget) qtPaintsCard() bool { return true }

func (w *AcrylicWidget) paint() {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	w.drawAcrylicContent(painter, 0, 0, false)
	painter.End()
}

// drawAcrylic paints this card's acrylic content (rounded, blurred backdrop +
// tint + children + border) at the given offset, for the sibling compositing in
// buildWindowComposite.
func (w *AcrylicWidget) drawAcrylic(painter *qt.QPainter, x, y int) {
	w.drawAcrylicContent(painter, x, y, true)
}

// drawAcrylicContent paints the acrylic itself. withChildren controls whether the
// card's child widgets are included: yes when compositing this card below
// another one, no when Qt is about to paint them over the card anyway.
func (w *AcrylicWidget) drawAcrylicContent(painter *qt.QPainter, x, y int, withChildren bool) {
	width := w.Width()
	height := w.Height()

	painter.Save()
	painter.SetRenderHint2(qt.QPainter__Antialiasing, true)
	painter.SetRenderHint2(qt.QPainter__SmoothPixmapTransform, true)

	// Clip to the rounded rectangle so the corners stay transparent.
	if w.cornerRadius > 0 {
		path := qt.NewQPainterPath()
		path.AddRoundedRect2(float64(x), float64(y), float64(width), float64(height),
			float64(w.cornerRadius), float64(w.cornerRadius))
		painter.SetClipPath(path)
		path.Delete()
	}

	if w.cached != nil {
		rect := qt.NewQRect4(x, y, width, height)
		painter.DrawImage6(rect, w.cached)
		rect.Delete()
	} else {
		// Backdrop not captured yet; paint an opaque base so the widget does
		// not show stale garbage behind it.
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

	// Child widgets are painted over the acrylic, so they belong to the picture
	// other cards see when they composite this one. The card's own paintEvent
	// leaves them out: Qt paints its children right after it.
	if withChildren {
		renderChildWidgets(w.QWidget, painter, x, y)
	}

	painter.Restore()

	// Border (drawn unclipped so it can sit exactly on the rounded edge).
	if w.hasBorder && w.borderWidth > 0 {
		colour := qt.NewQColor11(w.borderR, w.borderG, w.borderB, w.borderA)
		pen := qt.NewQPen3(colour)
		pen.SetWidth(w.borderWidth)
		painter.SetPenWithPen(pen)
		painter.SetBrushWithStyle(qt.NoBrush)

		// Inset by half the pen width so the stroke stays inside the widget.
		inset := float64(w.borderWidth) / 2.0
		r := float64(w.cornerRadius)
		if r > 0 {
			r -= inset
			if r < 0 {
				r = 0
			}
		}
		painter.DrawRoundedRect2(
			x+int(inset), y+int(inset),
			width-w.borderWidth, height-w.borderWidth, r, r)

		pen.Delete()
		colour.Delete()
	}

	if w.caption != "" {
		font := captionFont(w.QWidget)
		painter.SetFont(font)
		font.Delete()

		textColor := qt.NewQColor11(255, 255, 255, 255)
		painter.SetPen(textColor)
		textColor.Delete()

		pos := qt.NewQPointF3(float64(x+16), float64(y+30))
		painter.DrawText(pos, w.caption)
		pos.Delete()
	}
}

// drawCaptionWithPainter paints a card's caption with a QPainter on the card
// widget, using an explicitly sized font (see captionFont).
//
// Only the *non-native* back end (AcrylicOpenGLWidget) may use this. On a
// QGLWidget, beginning a QPainter on the native GL surface disturbs it — the card
// then showed the widget's palette colour instead of the acrylic. That back end
// uses QGLWidget::renderText with the same font instead, which renders the text
// offscreen and only blits a texture.
func drawCaptionWithPainter(w *qt.QWidget, caption string) {
	if w == nil || caption == "" {
		return
	}

	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()

	font := captionFont(w)
	painter.SetFont(font)
	font.Delete()

	colour := qt.NewQColor11(255, 255, 255, 255)
	painter.SetPen(colour)
	colour.Delete()

	at := qt.NewQPoint2(16, 30)
	painter.DrawText2(at, caption)
	at.Delete()

	painter.End()
}

// captionFont builds the font used for a card's caption.
//
// The point size is converted to an explicit pixel size using the widget's own
// DPI, because Qt's OpenGL paint engine resolves point sizes against a fixed
// 96 DPI: on a 144 DPI screen the GPU back ends therefore drew their captions at
// two thirds of the size the CPU back end used, which looked like broken text.
func captionFont(w *qt.QWidget) *qt.QFont {
	font := qt.NewQFont2("Segoe UI")
	font.SetPointSize(11)

	if w != nil {
		if dpi := w.QPaintDevice.LogicalDpiY(); dpi > 0 {
			if px := int(float64(font.PointSizeF())*float64(dpi)/72.0 + 0.5); px > 0 {
				font.SetPixelSize(px)
			}
		}
	}
	return font
}

