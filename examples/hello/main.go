// Command hello is the canonical gofluent example template. It demonstrates the
// three conventions every example follows:
//
//  1. Use gofluent/examples/internal/demo for the shared entry point.
//  2. Embed example assets with go:embed (resource embedding convention).
//  3. Keep each example a self-contained `package main` under examples/<category>/<name>.
package main

import (
	_ "embed"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

//go:embed resource/hello.qss
var helloQSS string

func main() {
	// Optionally pin the theme; ThemeAuto follows the OS (and degrades to light
	// on hosts without a system-theme signal, per MIGRATION_GUIDE §8).
	common.SetTheme(common.ThemeAuto, false, false)

	demo.Run(func() *qt.QWidget {
		w := qt.NewQWidget2()
		w.SetObjectName("helloWindow")
		w.Resize(360, 140)
		w.SetStyleSheet(helloQSS)

		layout := qt.NewQVBoxLayout(w)
		label := qt.NewQLabel3("Hello from gofluent / miqt")
		label.SetAlignment(qt.AlignCenter)
		layout.AddWidget(label.QWidget)
		return w
	})
}
