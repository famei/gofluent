// Package demo provides the shared application entry point used by every
// gofluent example. It centralizes the High-DPI setup, QApplication creation,
// and event-loop startup that the upstream PyQt demos repeat in each
// `if __name__ == '__main__'` block.
//
// Simple demo:
//
//	package main
//
//	import "gofluent/examples/internal/demo"
//
//	func main() {
//		demo.Run(newDemoWidget)
//	}
//
// For examples that must configure the application before the event loop starts
// (install translators, set the application icon, decide the DPI policy from
// config, ...), use the lower-level SetupHighDPI / NewApp / Exec steps instead
// of Run — see examples/gallery.
package demo

import (
	"os"

	qt "github.com/mappu/miqt/qt"
)

// app keeps the QApplication object reachable for the whole process lifetime.
// miqt locks the OS thread on the first NewQApplication call, so every function
// in this package must be invoked from the main goroutine.
var app *qt.QApplication

// SetupHighDPI applies the shared High-DPI attributes used by the upstream
// PyQt demos:
//
//	QApplication.setHighDpiScaleFactorRoundingPolicy(PassThrough)
//	QApplication.setAttribute(AA_EnableHighDpiScaling)
//	QApplication.setAttribute(AA_UseHighDpiPixmaps)
//
// Call it before NewApp only when configuration must be read before the
// application object exists. Run calls it automatically.
func SetupHighDPI() {
	qt.QGuiApplication_SetHighDpiScaleFactorRoundingPolicy(qt.PassThrough)
	qt.QCoreApplication_SetAttribute(qt.AA_EnableHighDpiScaling)
	qt.QCoreApplication_SetAttribute(qt.AA_UseHighDpiPixmaps)
}

// NewApp creates the single QApplication for the process and keeps a package
// level reference so the object is never garbage collected. Call it from the
// main goroutine only.
func NewApp() *qt.QApplication {
	app = qt.NewQApplication(os.Args)
	return app
}

// Exec runs the Qt event loop and exits the process with its return code.
// It never returns.
func Exec() {
	os.Exit(qt.QApplication_Exec())
}

// Run is the one-line entry point for demos that do not need custom
// pre-event-loop setup. It applies the High-DPI settings, creates the
// QApplication, then calls each supplied factory to construct a top-level
// widget (so QWidgets are only created after the QApplication exists), shows
// each result, and runs the event loop until quit.
//
// Each factory has signature `func() *qt.QWidget` and is invoked in order after
// NewApp, e.g.:
//
//	demo.Run(newDemo)
func Run(newWindows ...func() *qt.QWidget) {
	SetupHighDPI()
	NewApp()
	for _, newWindow := range newWindows {
		newWindow().Show()
	}
	Exec()
}
