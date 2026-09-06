// Package asset provides helpers for reading example assets embedded with
// go:embed and converting them into the miqt value objects (QImage, QPixmap,
// QIcon, fonts) that the demos consume.
//
// Convention: each example that ships assets places them in a `resource/`
// subdirectory next to its `main.go` and embeds them with go:embed:
//
//	import _ "embed"
//
//	//go:embed resource/logo.png
//	var logoPNG []byte
//
//	//go:embed resource/dark/*.qss
//	var darkQSS embed.FS
//
// go:embed cannot reference files outside the package directory, so assets must
// be copied into each example's own directory tree (see examples/README.md).
package asset

import (
	"fmt"
	"io/fs"

	qt "github.com/mappu/miqt/qt"
)

// Bytes reads an embedded file. It panics on error because example assets are
// baked in at build time; a missing file is a programming error that should
// surface immediately rather than render a broken demo.
func Bytes(fsys fs.FS, name string) []byte {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		panic(fmt.Sprintf("asset: read %s: %v", name, err))
	}
	return b
}

// String reads an embedded text file (QSS, translation, JSON, ...).
func String(fsys fs.FS, name string) string {
	return string(Bytes(fsys, name))
}

// QImage decodes an embedded image (PNG/JPG/SVG via Qt's image plugins) into a
// *qt.QImage. The caller owns the returned object and should Delete() it when
// done (see MIGRATION_GUIDE §6).
func QImage(fsys fs.FS, name string) *qt.QImage {
	img := qt.QImage_FromDataWithData(Bytes(fsys, name))
	if img == nil || img.IsNull() {
		panic(fmt.Sprintf("asset: decode image %s", name))
	}
	return img
}

// QPixmap decodes an embedded image into a *qt.QPixmap. The caller owns the
// returned object and should Delete() it when done.
func QPixmap(fsys fs.FS, name string) *qt.QPixmap {
	pm := qt.NewQPixmap()
	if !pm.LoadFromDataWithData(Bytes(fsys, name)) {
		pm.Delete()
		panic(fmt.Sprintf("asset: decode pixmap %s", name))
	}
	return pm
}

// QIcon decodes an embedded image into a *qt.QIcon. The caller owns the
// returned object and should Delete() it when done.
func QIcon(fsys fs.FS, name string) *qt.QIcon {
	pm := QPixmap(fsys, name)
	defer pm.Delete()
	return qt.NewQIcon2(pm)
}

// FontID loads an embedded font file and returns the font ID assigned by Qt
// (see QFontDatabase.AddApplicationFontFromData). A negative return value means
// the font could not be loaded.
func FontID(data []byte) int {
	return qt.QFontDatabase_AddApplicationFontFromData(data)
}
