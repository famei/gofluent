// Package dialog_box ports qfluentwidgets/components/dialog_box to Go + miqt.
//
// It provides the mask dialog base (MaskDialogBase), the plain and message
// dialogs (Dialog, MessageBox, MessageDialog, MessageBoxBase), the color picker
// dialog (ColorDialog) and the folder list dialog (FolderListDialog).
//
// Dependencies (all strictly one-way):
//   - gofluent/common     (config, style sheets, text wrap)
//   - gofluent/components/widgets (buttons, labels, line edits, scroll areas, sliders)
//
// Signals from the Python source (pyqtSignal) are expressed as Go callback
// fields (OnYes, OnCancel, OnColorChanged, OnFolderChanged, ...). The fade
// in/out opacity animation in MaskDialogBase is retained via QPropertyAnimation
// on the "opacity" property of a QGraphicsOpacityEffect.
package dialog_box
