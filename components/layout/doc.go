// Package layout ports the qfluentwidgets/components/layout package (4 Python
// source files) from PyQt5 to Go on top of github.com/mappu/miqt (Qt5).
//
// It provides the fluent layout primitives used by the setting cards and
// general container widgets: ExpandLayout (vertical content-driven stacking),
// FlowLayout / AdaptiveFlowLayout (flowing rows) and VBoxLayout (a small
// convenience wrapper around QVBoxLayout).
//
// Dependency direction: layout depends on gofluent/common; it does not import
// sibling component packages to avoid cycles.
package layout
