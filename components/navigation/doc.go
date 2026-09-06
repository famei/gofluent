// Package navigation ports the qfluentwidgets/components/navigation package
// (7 Python source files) from PyQt5 to Go on top of github.com/mappu/miqt
// (Qt5 bindings).
//
// It depends on gofluent/common and gofluent/components/widgets (buttons, tool
// tips, scroll areas, flyouts and menus) and reuses the shared QSS resources
// through gofluent/resources. The FluentStyleSheet(NAVIGATION_INTERFACE/PIVOT)
// styles are applied via SetStyleSheet exactly like the Python port.
//
// Custom meta-property animations (ScaleSlideAnimation, IconSlideAnimation,
// radius/textOpacity, maximumHeight) are reproduced with the common
// ProgressAnimation driver (see animation.go) because miqt cannot register
// arbitrary Qt meta-properties from Go. The acrylic flyout remains simplified
// (see MIGRATION_GUIDE §10).
package navigation
