// Command flip_view migrates examples/view/flip_view/demo.py.
package main

import (
	"embed"
	"io/fs"
	"path"
	"strings"

	"github.com/famei/gofluent/components/widgets"
	"github.com/famei/gofluent/examples/internal/asset"
	"github.com/famei/gofluent/examples/internal/demo"
	qt "github.com/mappu/miqt/qt"
)

//go:embed resource/*
var imageFS embed.FS

func newDemo() *qt.QWidget {
	w := qt.NewQWidget2()

	flipView := widgets.NewHorizontalFlipView(w)
	pager := widgets.NewHorizontalPipsPager(w)

	flipView.SetAspectRatioMode(qt.KeepAspectRatio)

	names, err := fs.Glob(imageFS, "resource/*")
	if err != nil {
		panic(err)
	}
	images := make([]interface{}, 0, len(names))
	for _, name := range names {
		// Skip hidden files (e.g. the ".gitkeep" placeholder) which are not images.
		if strings.HasPrefix(path.Base(name), ".") {
			continue
		}
		images = append(images, asset.QImage(imageFS, name))
	}
	flipView.AddImages(images)
	pager.SetPageNumber(flipView.Count())

	pager.OnCurrentIndexChanged(func(index int) { flipView.SetCurrentIndex(index) })
	flipView.OnCurrentIndexChanged(func(index int) { pager.SetCurrentIndex(index) })

	layout := qt.NewQVBoxLayout(w)
	layout.AddWidget3(flipView.QWidget, 0, qt.AlignCenter)
	layout.AddWidget3(pager.QWidget, 0, qt.AlignCenter)
	layout.SetSpacing(20)
	w.Resize(600, 600)

	return w
}

func main() {
	demo.Run(newDemo)
}
