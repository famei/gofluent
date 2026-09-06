// Package resource embeds the PNG images used by the clock demo. The upstream
// resource.qrc compiled these into the binary with pyrcc; the Go port loads the
// original files through go:embed instead (see examples/README.md §3).
package resource

import (
	"embed"

	qt "github.com/mappu/miqt/qt"

	"github.com/famei/gofluent/examples/internal/asset"
)

//go:embed images/*.png
var FS embed.FS

// Tips returns the tips.png pixmap used by the progress card icon.
func Tips() *qt.QPixmap { return asset.QPixmap(FS, "images/tips.png") }

// Alarms returns the alarms.png pixmap used by the focus card icon.
func Alarms() *qt.QPixmap { return asset.QPixmap(FS, "images/alarms.png") }

// Todo returns the todo.png pixmap used by the task card icon.
func Todo() *qt.QPixmap { return asset.QPixmap(FS, "images/todo.png") }

// Shoko returns the shoko.png pixmap used by the navigation avatar.
func Shoko() *qt.QPixmap { return asset.QPixmap(FS, "images/shoko.png") }
