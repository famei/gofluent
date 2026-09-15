package widgets

// widgetAlive tracks whether a widget is still alive, for callbacks that are
// connected to a QAction instead of to the widget itself.
//
// Actions are shared and commonly outlive the widgets that listen to them: they
// are created without a parent and handed to menus and command bars, while a
// widget tree is deleted by its parent (a command-bar flyout deletes its whole
// bar — and every CommandButton — when it closes). Without this guard the
// action's changed/toggled signals call into freed widget memory. Switching the
// theme is the classic trigger: it re-renders every action icon, which fires
// changed() into the already destroyed buttons and crashes the process.
type widgetAlive struct {
	alive bool
}

// trackWidget starts tracking the lifetime of a widget. onDestroyed is the
// widget's own OnDestroyed registration method (e.g. b.OnDestroyed).
func trackWidget(onDestroyed func(func())) *widgetAlive {
	a := &widgetAlive{alive: true}
	onDestroyed(func() { a.alive = false })
	return a
}

// ok reports whether the tracked widget is still alive.
func (a *widgetAlive) ok() bool { return a != nil && a.alive }
