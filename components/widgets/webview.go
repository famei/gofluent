package widgets

import (
	"sync"
	"unsafe"

	"github.com/crgimenes/glaze"
	"github.com/mappu/miqt/qt"
)

// DefaultInitDelay is how long the widget's geometry is given to settle before the
// embedded browser is created (see SetInitDelay).
const DefaultInitDelay = 250

type QWebEngineView struct {
	*qt.QWidget
	Webview glaze.WebView
	once    sync.Once

	// Requests made before the embedded browser exists (see ensureWebview).
	pendingURL  string
	pendingHTML string

	// initTimer defers creation until the geometry has stopped changing.
	initTimer *qt.QTimer
	initDelay int
	initFn    func()
}

// NewQWebEngineView There is a potential risk of Qt rendering errors when using this control. 🤯🤯🤯
func NewQWebEngineView(parent *qt.QWidget) *QWebEngineView {
	s := &QWebEngineView{
		QWidget:   qt.NewQWidget(parent),
		initDelay: DefaultInitDelay,
	}

	// The embedded browser sizes its own native window to the host HWND when it is
	// created, so it must be created only once the widget has its *final* geometry.
	// Neither the constructor (a widget that has not been laid out reports its
	// default 100x30) nor showEvent is late enough: the layouts, the stacked widget
	// and the Mica setup still change the size after that, and the page then renders
	// at the wrong size/position until something forces a resize. Creation is
	// therefore deferred by a short timer that every resize restarts, so it happens
	// once the geometry has settled.
	s.initTimer = qt.NewQTimer2(s.QObject)
	s.initTimer.SetSingleShot(true)
	s.initTimer.SetInterval(s.initDelay)
	s.initTimer.OnTimeout(func() { s.ensureWebview() })

	s.OnResizeEvent(func(super func(e *qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		s.scheduleWebview()
	})
	s.OnShowEvent(func(super func(e *qt.QShowEvent), e *qt.QShowEvent) {
		super(e)
		s.scheduleWebview()
	})

	s.OnDestroyed(func() {
		s.once.Do(func() {
			if s.Webview != nil {
				s.Webview.Destroy()
			}
		})
	})
	return s
}

func (w *QWebEngineView) OnInit(f func()) {
	w.initFn = f
}

// SetInitDelay changes how long the geometry is given to settle before the embedded
// browser is created (milliseconds; DefaultInitDelay by default). It has no effect
// once the browser exists.
func (w *QWebEngineView) SetInitDelay(milliseconds int) {
	if milliseconds < 0 {
		milliseconds = 0
	}
	w.initDelay = milliseconds
	if w.initTimer != nil {
		w.initTimer.SetInterval(milliseconds)
	}
}

// scheduleWebview (re)arms the settle timer. Every resize restarts it while the
// browser does not exist yet, so creation waits for the geometry to stop changing
// (a widget that is not shown or has no size yet arms nothing; the show/resize that
// gives it one does).
func (w *QWebEngineView) scheduleWebview() {
	if w.Webview != nil || w.initTimer == nil {
		return
	}
	if !w.IsVisible() || w.Width() <= 1 || w.Height() <= 1 {
		return
	}
	w.initTimer.Start(w.initDelay)
}

// ensureWebview creates the embedded browser once the widget has a usable size, and
// replays any URL/HTML requested before it existed.
func (w *QWebEngineView) ensureWebview() {
	if w.Webview != nil || w.Width() <= 1 || w.Height() <= 1 {
		return
	}
	wv, err := glaze.NewWindow(true, unsafe.Pointer(w.WinId()))
	if err != nil {
		return
	}
	w.Webview = wv
	if w.pendingURL != "" {
		w.Webview.Navigate(w.pendingURL)
	}
	if w.pendingHTML != "" {
		w.Webview.SetHtml(w.pendingHTML)
	}
	if w.initFn != nil {
		w.initFn()
	}
}

// IsReady reports whether the embedded browser has been created.
func (w *QWebEngineView) IsReady() bool { return w.Webview != nil }

func (w *QWebEngineView) SetUrl(url string) {
	if w.Webview == nil {
		w.pendingURL = url
		return
	}
	w.Webview.Navigate(url)
}

func (w *QWebEngineView) SetHtml(html string) {
	if w.Webview == nil {
		w.pendingHTML = html
		return
	}
	w.Webview.SetHtml(html)
}

func (w *QWebEngineView) Eval(js string) {
	if w.Webview == nil {
		return
	}
	w.Webview.Dispatch(func() {
		w.Webview.Eval(js)
	})
}

func (w *QWebEngineView) Bind(name string, f any) error {
	if w.Webview == nil {
		return nil
	}
	return w.Webview.Bind(name, f)
}

func (w *QWebEngineView) Unbind(name string) error {
	if w.Webview == nil {
		return nil
	}
	return w.Webview.Unbind(name)
}
