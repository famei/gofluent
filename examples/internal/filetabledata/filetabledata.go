// Package filetabledata builds the fake directory listing used by the FileTable
// demo (examples/view/file_table) and by the gallery sample, so both show the
// same Windows Explorer like content.
package filetabledata

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// FolderCount is the number of directories placed at the top of the listing.
const FolderCount = 12

// Columns of the demo listing.
var headers = []string{"名称", "修改日期", "类型", "大小"}

var folderNames = []string{
	"3D Objects", "Documents", "Downloads", "Music", "Pictures", "Videos",
	"gofluent", "Qt", "Qt5", "PyQt-Fluent-Widgets", "build", "resource",
}

type fileKind struct {
	extension string
	typeName  string
	maxSize   int
}

var fileKinds = []fileKind{
	{"txt", "文本文档", 512},
	{"md", "Markdown 文档", 128},
	{"go", "Go 源文件", 256},
	{"json", "JSON 文件", 2048},
	{"png", "PNG 图片", 8192},
	{"jpg", "JPG 图片", 6144},
	{"svg", "SVG 矢量图", 64},
	{"mp4", "MP4 视频", 1024 * 512},
	{"zip", "压缩文件", 1024 * 256},
}

var namePrefixes = []string{
	"report", "design", "screenshot", "notes", "backup", "avatar", "icon",
	"main", "widget", "animation", "navigation", "table", "theme", "config",
	"README", "CHANGELOG", "LICENSE", "demo", "sample", "example",
}

// Fill fills table with a directory listing of fileCount files preceded by
// FolderCount folders, sets the headers, the column widths and the icons.
func Fill(table *widgets.FileTable, fileCount int) {
	table.SetColumnCount(len(headers))
	table.SetHorizontalHeaderLabels(headers)
	table.SetRowCount(FolderCount + fileCount)

	rng := rand.New(rand.NewSource(20260920))
	for row := 0; row < FolderCount; row++ {
		table.SetItem(row, 0, qt.NewQTableWidgetItem2(folderNames[row%len(folderNames)]))
		table.SetItem(row, 1, qt.NewQTableWidgetItem2(fmt.Sprintf("2026/09/%02d %02d:%02d", 1+row%28, row%24, row%60)))
		table.SetItem(row, 2, qt.NewQTableWidgetItem2("文件夹"))
		table.SetItem(row, 3, qt.NewQTableWidgetItem2(""))
	}

	for i := 0; i < fileCount; i++ {
		kind := fileKinds[rng.Intn(len(fileKinds))]
		row := FolderCount + i
		name := fmt.Sprintf("%s_%03d.%s", namePrefixes[i%len(namePrefixes)], i, kind.extension)
		size := 1 + rng.Intn(kind.maxSize)
		table.SetItem(row, 0, qt.NewQTableWidgetItem2(name))
		table.SetItem(row, 1, qt.NewQTableWidgetItem2(fmt.Sprintf("2026/%02d/%02d %02d:%02d", 1+rng.Intn(9), 1+rng.Intn(28), rng.Intn(24), rng.Intn(60))))
		table.SetItem(row, 2, qt.NewQTableWidgetItem2(kind.typeName))
		table.SetItem(row, 3, qt.NewQTableWidgetItem2(formatSize(size)))
	}

	table.SetColumnWidth(0, 240)
	table.SetColumnWidth(1, 150)
	table.SetColumnWidth(2, 130)
	table.SetColumnWidth(3, 110)
	table.SetMaxColumn(len(headers) - 1)

	ApplyIcons(table)
	WatchIcons(table)
}

// ApplyIcons (re)renders the folder/document glyph of every row. The glyphs are
// drawn with the theme ink color, so they have to be re-rendered after a theme
// switch, otherwise the black icons stay invisible on a dark background.
func ApplyIcons(table *widgets.FileTable) {
	folderIcon := common.ToQIcon(common.Folder)
	fileIcon := common.ToQIcon(common.Document)
	defer folderIcon.Delete()
	defer fileIcon.Delete()

	for row := 0; row < table.RowCount(); row++ {
		item := table.Item(row, 0)
		if item == nil {
			continue
		}
		if row < FolderCount {
			item.SetIcon(folderIcon)
		} else {
			item.SetIcon(fileIcon)
		}
	}
}

// WatchIcons re-applies the row icons on every theme switch.
func WatchIcons(table *widgets.FileTable) {
	state := &iconState{alive: true}
	table.OnDestroyed(func() { state.alive = false })
	common.QConfigInstance.OnThemeChanged(func(common.Theme) {
		if !state.alive {
			return
		}
		ApplyIcons(table)
		table.Viewport().Update()
	})
}

type iconState struct {
	alive bool
}

// HighlightMatching returns the rows whose name contains text (case
// insensitive), for table.SetHighlight.
func HighlightMatching(table *widgets.FileTable, text string) map[int]struct{} {
	rows := make(map[int]struct{})
	needle := strings.ToLower(strings.TrimSpace(text))
	if needle == "" {
		return rows
	}
	for row := 0; row < table.RowCount(); row++ {
		item := table.Item(row, 0)
		if item == nil {
			continue
		}
		if strings.Contains(strings.ToLower(item.Text()), needle) {
			rows[row] = struct{}{}
		}
	}
	return rows
}

// formatSize renders a size in KB like Windows Explorer does.
func formatSize(kb int) string {
	if kb >= 1024 {
		return fmt.Sprintf("%d,%03d KB", kb/1024, kb%1024)
	}
	return fmt.Sprintf("%d KB", kb)
}
