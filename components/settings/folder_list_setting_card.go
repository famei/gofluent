package settings

import (
	"path/filepath"

	"github.com/famei/gofluent/common"
	dialog_box "github.com/famei/gofluent/components/dialog_box"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
)

// FolderItem is a single folder row with a remove button.
type FolderItem struct {
	*qt.QWidget
	folder       string
	hBoxLayout   *qt.QHBoxLayout
	folderLabel  *qt.QLabel
	removeButton *widgets.ToolButton

	OnRemoved func(item *FolderItem)
}

// NewFolderItem builds a folder item.
func NewFolderItem(folder string, parent *qt.QWidget) *FolderItem {
	item := &FolderItem{QWidget: qt.NewQWidget(parent), folder: folder}
	item.hBoxLayout = qt.NewQHBoxLayout(item.QWidget)
	item.folderLabel = qt.NewQLabel5(folder, item.QWidget)
	item.removeButton = widgets.NewToolButtonIcon(common.Close, item.QWidget)

	item.removeButton.SetFixedSize2(39, 29)
	item.removeButton.SetIconSize(qt.NewQSize2(12, 12))

	item.SetFixedHeight(53)
	item.SetSizePolicy2(qt.QSizePolicy__Ignored, qt.QSizePolicy__Fixed)
	item.hBoxLayout.SetContentsMargins(48, 0, 60, 0)
	item.hBoxLayout.AddWidget3(item.folderLabel.QWidget, 0, qt.AlignLeft)
	item.hBoxLayout.AddSpacing(16)
	item.hBoxLayout.AddStretchWithStretch(1)
	item.hBoxLayout.AddWidget3(item.removeButton.QWidget, 0, qt.AlignRight)

	item.folderLabel.SetObjectName("titleLabel")

	item.removeButton.OnClicked(func() {
		if item.OnRemoved != nil {
			item.OnRemoved(item)
		}
	})
	return item
}

// Folder returns the folder path.
func (i *FolderItem) Folder() string { return i.folder }

// FolderListSettingCard is an expandable card listing folders with an add
// button in the header.
type FolderListSettingCard struct {
	*ExpandSettingCard
	configItem      *common.ConfigItem
	dialogDirectory string
	addFolderButton *widgets.PushButton
	folders         []string

	OnFolderChanged func(folders []string)
}

// NewFolderListSettingCard builds a folder list setting card.
func NewFolderListSettingCard(configItem *common.ConfigItem, title, content, directory string, parent *qt.QWidget) *FolderListSettingCard {
	c := &FolderListSettingCard{ExpandSettingCard: NewExpandSettingCard(common.Folder, title, content, parent)}
	c.configItem = configItem
	c.dialogDirectory = directory
	c.addFolderButton = widgets.NewPushButtonIcon(common.FolderAdd, "Add folder", c.QWidget)

	c.folders = append([]string(nil), asStringSlice(configItem.Value())...)
	c.initWidget()
	return c
}

func (c *FolderListSettingCard) initWidget() {
	c.AddWidget(c.addFolderButton.QWidget)

	c.viewLayout.SetSpacing(0)
	c.viewLayout.SetContentsMargins(0, 0, 0, 0)
	for _, folder := range c.folders {
		c.addFolderItem(folder)
	}

	c.addFolderButton.OnClicked(c.showFolderDialog)
}

func (c *FolderListSettingCard) showFolderDialog() {
	folder := qt.QFileDialog_GetExistingDirectory3(c.QWidget, "Choose folder", c.dialogDirectory)
	if folder == "" {
		return
	}
	for _, f := range c.folders {
		if f == folder {
			return
		}
	}

	c.addFolderItem(folder)
	c.folders = append(c.folders, folder)
	common.QConfigInstance.Set(c.configItem, c.folders, true)
	if c.OnFolderChanged != nil {
		c.OnFolderChanged(c.folders)
	}
}

func (c *FolderListSettingCard) addFolderItem(folder string) {
	item := NewFolderItem(folder, c.view.QWidget)
	item.OnRemoved = func(item *FolderItem) { c.showConfirmDialog(item) }
	c.viewLayout.AddWidget(item.QWidget)
	item.Show()
	c.adjustViewSize()
}

func (c *FolderListSettingCard) showConfirmDialog(item *FolderItem) {
	name := filepath.Base(item.folder)
	title := "Are you sure you want to delete the folder?"
	content := "If you delete the \"" + name + "\" folder and remove it from the list, the folder will no longer appear in the list, but will not be deleted."
	d := dialog_box.NewDialog(title, content, c.Window())
	d.OnYes = func() { c.removeFolder(item) }
	d.Exec()
}

func (c *FolderListSettingCard) removeFolder(item *FolderItem) {
	index := -1
	for i, f := range c.folders {
		if f == item.folder {
			index = i
			break
		}
	}
	if index < 0 {
		return
	}

	c.folders = append(c.folders[:index], c.folders[index+1:]...)
	c.viewLayout.RemoveWidget(item.QWidget)
	item.DeleteLater()
	c.adjustViewSize()

	if c.OnFolderChanged != nil {
		c.OnFolderChanged(c.folders)
	}
	common.QConfigInstance.Set(c.configItem, c.folders, true)
}

// Folders returns the current folder list.
func (c *FolderListSettingCard) Folders() []string { return c.folders }

func asStringSlice(v interface{}) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, e := range x {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
