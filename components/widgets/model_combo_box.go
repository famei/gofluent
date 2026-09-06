package widgets

import (
	"github.com/famei/gofluent/common"
	qt "github.com/mappu/miqt/qt"
)

// ModelComboBox is the model-backed combo box. The Go port keeps the same
// []ComboItem backing as ComboBox (the QStandardItemModel is simplified away)
// and adds the button icon visibility feature.
type ModelComboBox struct {
	*ComboBox
	isIconVisible bool
}

// NewModelComboBox builds a model combo box.
func NewModelComboBox(parent *qt.QWidget) *ModelComboBox {
	w := &ModelComboBox{ComboBox: NewComboBox(parent), isIconVisible: true}
	w.SetObjectName("modelComboBox")
	return w
}

// SetIconVisible toggles the visibility of the current item icon.
func (w *ModelComboBox) SetIconVisible(isVisible bool) {
	if isVisible == w.isIconVisible || w.CurrentIndex() < 0 {
		return
	}
	w.isIconVisible = isVisible
	if isVisible {
		w.updateIcon()
	} else {
		w.SetIcon(qt.NewQIcon())
	}
}

// IsIconVisible reports whether the current item icon is visible.
func (w *ModelComboBox) IsIconVisible() bool { return w.isIconVisible }

func (w *ModelComboBox) updateIcon() {
	if !w.isIconVisible {
		return
	}
	icon := w.ItemIcon(w.CurrentIndex())
	if icon != nil {
		w.SetIcon(common.ToQIcon(icon))
	} else {
		w.SetIcon(qt.NewQIcon())
	}
}

// SetCurrentIndex sets the current index and refreshes the button icon.
func (w *ModelComboBox) SetCurrentIndex(index int) {
	w.ComboBox.SetCurrentIndex(index)
	w.updateIcon()
}

// Clear removes all items and resets the selection.
func (w *ModelComboBox) Clear() {
	w.ComboBox.Clear()
	w.SetCurrentIndex(-1)
}

// SetItemIcon sets an item icon and refreshes the button icon when current.
func (w *ModelComboBox) SetItemIcon(index int, icon interface{}) {
	w.ComboBox.SetItemIcon(index, icon)
	if index == w.CurrentIndex() && w.isIconVisible {
		w.SetIcon(common.ToQIcon(icon))
	}
}

// EditableModelComboBox is the model-backed editable combo box. It mirrors
// EditableComboBox exactly (the model layer is collapsed to []ComboItem).
type EditableModelComboBox struct {
	*EditableComboBox
}

// NewEditableModelComboBox builds an editable model combo box.
func NewEditableModelComboBox(parent *qt.QWidget) *EditableModelComboBox {
	return &EditableModelComboBox{EditableComboBox: NewEditableComboBox(parent)}
}
