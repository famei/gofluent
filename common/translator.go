package common

import (
	"github.com/famei/gofluent/resources"
	qt "github.com/mappu/miqt/qt"
)

// FluentTranslator loads the embedded qfluentwidgets translations for a locale.
type FluentTranslator struct {
	*qt.QTranslator
	// data keeps the loaded ".qm" bytes alive: QTranslator::load(const uchar*,
	// int) does not copy the buffer, so the Go slice must outlive the translator
	// or the translation becomes dangling memory (same contract as
	// gallery.GalleryTranslator).
	data []byte
}

// NewFluentTranslator builds a translator; a nil locale falls back to the
// system locale.
func NewFluentTranslator(locale *qt.QLocale, parent *qt.QObject) *FluentTranslator {
	t := &FluentTranslator{QTranslator: qt.NewQTranslator2(parent)}
	if locale == nil {
		locale = qt.QLocale_System() // GoGC-armed — do NOT Delete
	}
	t.Load(locale)
	return t
}

// Load loads the embedded ".qm" file for the given locale into the translator.
func (t *FluentTranslator) Load(locale *qt.QLocale) bool {
	data := resources.Translation(resources.LocaleName(locale.Name()))
	if len(data) == 0 {
		return false
	}
	t.data = data
	return t.QTranslator.Load3(&data[0], len(data))
}
