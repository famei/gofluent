package common

import (
	"sync"

	qt "github.com/mappu/miqt/qt"
)

// RouteItem pairs a QStackedWidget with a route key.
type RouteItem struct {
	Stacked  *qt.QStackedWidget
	RouteKey string
}

// StackedHistory tracks the navigation history of one QStackedWidget.
type StackedHistory struct {
	stacked         *qt.QStackedWidget
	defaultRouteKey string
	history         []string
}

// NewStackedHistory builds a history for a stacked widget.
func NewStackedHistory(stacked *qt.QStackedWidget) *StackedHistory {
	h := &StackedHistory{stacked: stacked}
	h.history = []string{h.defaultRouteKey}
	return h
}

// Len returns the number of history entries.
func (h *StackedHistory) Len() int { return len(h.history) }

// IsEmpty reports whether the history only contains the default route.
func (h *StackedHistory) IsEmpty() bool { return len(h.history) <= 1 }

// Push appends routeKey if it differs from the current top.
func (h *StackedHistory) Push(routeKey string) bool {
	if len(h.history) > 0 && h.history[len(h.history)-1] == routeKey {
		return false
	}
	h.history = append(h.history, routeKey)
	return true
}

// Pop removes the top entry and navigates to the new top.
func (h *StackedHistory) Pop() {
	if h.IsEmpty() {
		return
	}
	h.history = h.history[:len(h.history)-1]
	h.goToTop()
}

// Remove deletes all occurrences of routeKey (after the default entry).
func (h *StackedHistory) Remove(routeKey string) {
	found := false
	for _, k := range h.history[1:] {
		if k == routeKey {
			found = true
			break
		}
	}
	if !found {
		return
	}
	rest := make([]string, 0, len(h.history)-1)
	rest = append(rest, h.history[0])
	for _, k := range h.history[1:] {
		if k != routeKey {
			rest = append(rest, k)
		}
	}
	h.history = dedupeStrings(rest)
	h.goToTop()
}

// Top returns the top route key.
func (h *StackedHistory) Top() string {
	if len(h.history) == 0 {
		return h.defaultRouteKey
	}
	return h.history[len(h.history)-1]
}

// SetDefaultRouteKey sets the default route and replaces the first entry.
func (h *StackedHistory) SetDefaultRouteKey(routeKey string) {
	h.defaultRouteKey = routeKey
	if len(h.history) == 0 {
		h.history = []string{routeKey}
	} else {
		h.history[0] = routeKey
	}
}

func (h *StackedHistory) goToTop() {
	if w := findWidgetByObjectName(h.stacked, h.Top()); w != nil {
		h.stacked.SetCurrentWidget(w)
	}
}

// Router tracks navigation history across multiple QStackedWidgets.
type Router struct {
	mu             sync.Mutex
	history        []RouteItem
	stackHistories map[*qt.QStackedWidget]*StackedHistory
	emptyListeners []func(bool)
}

// NewRouter builds an empty router.
func NewRouter() *Router {
	return &Router{stackHistories: map[*qt.QStackedWidget]*StackedHistory{}}
}

// OnEmptyChanged registers a listener for the emptyChanged signal.
func (r *Router) OnEmptyChanged(l func(bool)) {
	r.mu.Lock()
	r.emptyListeners = append(r.emptyListeners, l)
	r.mu.Unlock()
}

func (r *Router) emitEmpty() {
	r.mu.Lock()
	ls := append([]func(bool){}, r.emptyListeners...)
	r.mu.Unlock()
	for _, l := range ls {
		l(len(r.history) == 0)
	}
}

// SetDefaultRouteKey sets the default route key of a stacked widget.
func (r *Router) SetDefaultRouteKey(stacked *qt.QStackedWidget, routeKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.stackHistories[stacked]; !ok {
		r.stackHistories[stacked] = NewStackedHistory(stacked)
	}
	r.stackHistories[stacked].SetDefaultRouteKey(routeKey)
}

// Push pushes a route key onto a stacked widget's history.
func (r *Router) Push(stacked *qt.QStackedWidget, routeKey string) {
	r.mu.Lock()
	if _, ok := r.stackHistories[stacked]; !ok {
		r.stackHistories[stacked] = NewStackedHistory(stacked)
	}
	ok := r.stackHistories[stacked].Push(routeKey)
	if ok {
		r.history = append(r.history, RouteItem{Stacked: stacked, RouteKey: routeKey})
	}
	r.mu.Unlock()
	r.emitEmpty()
}

// Pop pops the most recent history entry.
func (r *Router) Pop() {
	r.mu.Lock()
	if len(r.history) == 0 {
		r.mu.Unlock()
		return
	}
	item := r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	r.mu.Unlock()

	r.emitEmpty()
	if h, ok := r.stackHistories[item.Stacked]; ok {
		h.Pop()
	}
}

// Remove removes all history entries with the given route key.
func (r *Router) Remove(routeKey string) {
	r.mu.Lock()
	filtered := r.history[:0]
	for _, i := range r.history {
		if i.RouteKey != routeKey {
			filtered = append(filtered, i)
		}
	}
	r.history = dedupeRouteItems(filtered)
	r.mu.Unlock()

	r.emitEmpty()

	for _, h := range r.stackHistories {
		if findWidgetByObjectName(h.stacked, routeKey) != nil {
			h.Remove(routeKey)
			return
		}
	}
}

// HistoryLen returns the number of history entries.
func (r *Router) HistoryLen() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.history)
}

// RouterInstance is the process-wide router (Python module-level `qrouter`).
var RouterInstance = NewRouter()

// findWidgetByObjectName returns the stacked widget page whose objectName is
// name, or nil. It replaces Python's stacked.findChild(QWidget, name).
func findWidgetByObjectName(stacked *qt.QStackedWidget, name string) *qt.QWidget {
	n := stacked.Count()
	for i := 0; i < n; i++ {
		w := stacked.Widget(i)
		if w != nil && w.ObjectName() == name {
			return w
		}
	}
	return nil
}

func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return in
	}
	out := in[:0]
	for i, s := range in {
		if i == 0 || s != in[i-1] {
			out = append(out, s)
		}
	}
	return out
}

func dedupeRouteItems(in []RouteItem) []RouteItem {
	if len(in) == 0 {
		return in
	}
	out := in[:0]
	for i, item := range in {
		if i == 0 || item.RouteKey != in[i-1].RouteKey {
			out = append(out, item)
		}
	}
	return out
}
