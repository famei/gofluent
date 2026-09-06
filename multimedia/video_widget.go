package multimedia

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/multimedia"
)

// GraphicsVideoItem is a QGraphicsVideoItem that paints with the difference
// composition mode (mirrors the Python GraphicsVideoItem.paint override).
type GraphicsVideoItem struct {
	*multimedia.QGraphicsVideoItem
}

// NewGraphicsVideoItem builds a graphics video item.
func NewGraphicsVideoItem() *GraphicsVideoItem {
	v := &GraphicsVideoItem{QGraphicsVideoItem: multimedia.NewQGraphicsVideoItem()}
	v.OnPaint(func(super func(painter *qt.QPainter, option *qt.QStyleOptionGraphicsItem, widget *qt.QWidget), painter *qt.QPainter, option *qt.QStyleOptionGraphicsItem, widget *qt.QWidget) {
		painter.SetCompositionMode(qt.QPainter__CompositionMode_Difference)
		super(painter, option, widget)
	})
	return v
}

// VideoWidget renders video through a QGraphicsView and overlays a standard
// play bar that fades in/out on hover.
type VideoWidget struct {
	*qt.QGraphicsView
	isHover       bool
	timer         *qt.QTimer
	vBoxLayout    *qt.QVBoxLayout
	videoItem     *GraphicsVideoItem
	graphicsScene *qt.QGraphicsScene
	playBar       *StandardMediaPlayBar
}

// NewVideoWidget builds a video widget.
func NewVideoWidget(parent *qt.QWidget) *VideoWidget {
	w := &VideoWidget{QGraphicsView: qt.NewQGraphicsView(parent)}
	w.isHover = false
	w.timer = qt.NewQTimer2(w.QObject)
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.videoItem = NewGraphicsVideoItem()
	w.graphicsScene = qt.NewQGraphicsScene4(w.QObject)
	w.playBar = NewStandardMediaPlayBar(w.QWidget)

	w.SetMouseTracking(true)
	w.SetScene(w.graphicsScene)
	w.graphicsScene.AddItem(w.videoItem.QGraphicsItem)
	w.SetVerticalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	w.SetRenderHints(qt.QPainter__Antialiasing | qt.QPainter__SmoothPixmapTransform)

	w.playBar.Player.SetVideoOutput(w.videoItem.QGraphicsVideoItem)
	common.FluentStyleSheet(common.FluentMediaPlayer).Apply(w.QWidget, common.ThemeAuto)

	w.timer.OnTimeout(w.onHideTimeOut)
	w.installEvents()
	return w
}

func (w *VideoWidget) installEvents() {
	w.OnHideEvent(func(super func(e *qt.QHideEvent), e *qt.QHideEvent) {
		w.Pause()
		e.Accept()
	})
	w.OnWheelEvent(func(super func(e *qt.QWheelEvent), e *qt.QWheelEvent) {
		// Swallow wheel events (Python wheelEvent returns immediately).
	})
	w.OnEnterEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = true
		w.playBar.FadeIn()
	})
	w.OnLeaveEvent(func(super func(e *qt.QEvent), e *qt.QEvent) {
		w.isHover = false
		w.timer.Start(3000)
	})
	w.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		sz := w.Size()
		sizeF := qt.NewQSizeF2(sz)
		w.videoItem.SetSize(sizeF)
		sizeF.Delete()

		w.FitInView5(w.videoItem.QGraphicsItem, qt.KeepAspectRatio)
		w.playBar.Move(11, w.Height()-w.playBar.Height()-11)
		w.playBar.SetFixedSize2(w.Width()-22, w.playBar.Height())
	})
}

// SetVideo sets the video source URL.
func (w *VideoWidget) SetVideo(url *qt.QUrl) {
	w.playBar.Player.SetSource(url)
	w.FitInView5(w.videoItem.QGraphicsItem, qt.KeepAspectRatio)
}

func (w *VideoWidget) onHideTimeOut() {
	if !w.isHover {
		w.playBar.FadeOut()
	}
}

// Play starts or resumes playback.
func (w *VideoWidget) Play() { w.playBar.Play() }

// Pause pauses playback.
func (w *VideoWidget) Pause() { w.playBar.Pause() }

// Stop stops playback.
func (w *VideoWidget) Stop() { w.playBar.Stop() }

// TogglePlayState toggles between play and pause.
func (w *VideoWidget) TogglePlayState() {
	if w.playBar.Player.IsPlaying() {
		w.Pause()
	} else {
		w.Play()
	}
}

// Player returns the underlying media player.
func (w *VideoWidget) Player() MediaPlayerBase { return w.playBar.Player }
