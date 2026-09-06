// Package common is the foundation of the gofluent module. It ports the
// qfluentwidgets/common package (theme/config system, stylesheets, icons,
// fonts, router, animation, scrolling and i18n) from PyQt5 to Go on top of
// github.com/mappu/miqt (Qt5 bindings).
//
// Dependency direction: common only depends on miqt and gofluent/resources;
// every other gofluent package (components/widgets, navigation, settings, ...)
// depends on common. Do not import any sibling package from here to avoid
// import cycles.
package common
