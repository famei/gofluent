// Package widgets ports the qfluentwidgets/components/widgets package (the
// largest module: 35 Python source files) from PyQt5 to Go on top of
// github.com/mappu/miqt (Qt5 bindings).
//
// Dependency direction: widgets depends only on gofluent/common and
// gofluent/resources; it must not import sibling packages to avoid cycles.
// The package exposes the fluent control set — buttons, labels, menus, tool
// tips, scroll bars, tab/table/tree/list views, sliders, spin boxes, switch
// buttons, progress indicators, info bar/badge, flyout, teaching tip,
// command bar, flip/pips pager, stacked widgets and the frameless window base.
package widgets
