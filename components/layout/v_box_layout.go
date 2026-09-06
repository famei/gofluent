package layout

import (
	qt "github.com/mappu/miqt/qt"
)

// VBoxLayout is a small convenience wrapper around QVBoxLayout that keeps
// track of every added widget so they can be removed or deleted in bulk.
type VBoxLayout struct {
	*qt.QVBoxLayout
	widgets []*qt.QWidget
}

// NewVBoxLayout builds a vertical box layout for parent.
func NewVBoxLayout(parent *qt.QWidget) *VBoxLayout {
	return &VBoxLayout{QVBoxLayout: qt.NewQVBoxLayout(parent)}
}

// AddWidget appends a widget (with stretch/alignment) and shows it.
func (l *VBoxLayout) AddWidget(widget *qt.QWidget, stretch int, alignment qt.AlignmentFlag) {
	l.QVBoxLayout.AddWidget3(widget, stretch, alignment)
	l.widgets = append(l.widgets, widget)
	widget.Show()
}

// AddWidgets appends several widgets.
func (l *VBoxLayout) AddWidgets(widgets []*qt.QWidget, stretch int, alignment qt.AlignmentFlag) {
	for _, w := range widgets {
		l.AddWidget(w, stretch, alignment)
	}
}

// RemoveWidget removes a widget from the layout without deleting it.
func (l *VBoxLayout) RemoveWidget(widget *qt.QWidget) {
	l.QVBoxLayout.RemoveWidget(widget)
	for i, w := range l.widgets {
		if w.UnsafePointer() == widget.UnsafePointer() {
			l.widgets = append(l.widgets[:i], l.widgets[i+1:]...)
			break
		}
	}
}

// DeleteWidget removes a widget from the layout and schedules it for deletion.
func (l *VBoxLayout) DeleteWidget(widget *qt.QWidget) {
	l.RemoveWidget(widget)
	widget.Hide()
	widget.DeleteLater()
}

// RemoveAllWidgets removes every tracked widget from the layout.
func (l *VBoxLayout) RemoveAllWidgets() {
	for _, w := range l.widgets {
		l.QVBoxLayout.RemoveWidget(w)
	}
	l.widgets = nil
}
