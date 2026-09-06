package common

import (
	"sync"

	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/examples/gallery/app/resource"
	qt "github.com/mappu/miqt/qt"
)

// Translator holds the localized navigation labels (port of
// app/common/translator.py). Each label is resolved through
// QCoreApplication.translate("Translator", ...) so the installed gallery .qm
// file translates them; the previous port stored the English source strings
// directly, which is why the gallery never switched to Chinese (bug #33).
type Translator struct {
	Text       string
	View       string
	Menus      string
	Icons      string
	Layout     string
	Dialogs    string
	Scroll     string
	Material   string
	DateTime   string
	Navigation string
	BasicInput string
	StatusInfo string
	Price      string
}

// Tr resolves a source string through the active Qt translation for the given
// context. It is the Go equivalent of QObject.tr(context, source) used across
// the Python gallery: the installed gallery ".qm" file maps (context, source)
// to the target language, so routing every user-visible string through Tr lets
// the whole UI follow the active locale. Contexts match the class names in the
// reference (SettingInterface, MainWindow, Translator, …).
func Tr(context, source string) string {
	return qt.QCoreApplication_Translate(context, source)
}

// NewTranslator builds a translator with the default labels resolved through
// the active translation ("Translator" context, matching translator.py).
func NewTranslator() *Translator {
	return &Translator{
		Text:       qt.QCoreApplication_Translate("Translator", "Text"),
		View:       qt.QCoreApplication_Translate("Translator", "View"),
		Menus:      qt.QCoreApplication_Translate("Translator", "Menus & toolbars"),
		Icons:      qt.QCoreApplication_Translate("Translator", "Icons"),
		Layout:     qt.QCoreApplication_Translate("Translator", "Layout"),
		Dialogs:    qt.QCoreApplication_Translate("Translator", "Dialogs & flyouts"),
		Scroll:     qt.QCoreApplication_Translate("Translator", "Scrolling"),
		Material:   qt.QCoreApplication_Translate("Translator", "Material"),
		DateTime:   qt.QCoreApplication_Translate("Translator", "Date & time"),
		Navigation: qt.QCoreApplication_Translate("Translator", "Navigation"),
		BasicInput: qt.QCoreApplication_Translate("Translator", "Basic input"),
		StatusInfo: qt.QCoreApplication_Translate("Translator", "Status & info"),
		Price:      qt.QCoreApplication_Translate("Translator", "Price"),
	}
}

// GalleryTranslator loads the embedded gallery ".qm" translations for a locale
// (the Go equivalent of galleryTranslator.load(locale, "gallery", ".", ":/gallery/i18n")).
type GalleryTranslator struct {
	*qt.QTranslator
	// data keeps the loaded ".qm" bytes alive: QTranslator::load(const uchar*,
	// int) does not copy the buffer, so the Go slice must outlive the
	// translator or the translation becomes dangling memory.
	data []byte
}

// NewGalleryTranslator builds a translator; a nil locale falls back to the
// system locale.
func NewGalleryTranslator(locale *qt.QLocale, parent *qt.QObject) *GalleryTranslator {
	t := &GalleryTranslator{QTranslator: qt.NewQTranslator2(parent)}
	if locale == nil {
		locale = qt.QLocale_System()
	}
	t.Load(locale)
	return t
}

// Load loads the embedded ".qm" file for the given locale.
func (t *GalleryTranslator) Load(locale *qt.QLocale) bool {
	data := resource.Translation(locale.Name())
	if len(data) == 0 {
		return false
	}
	t.data = data
	return t.QTranslator.Load3(&data[0], len(data))
}

// ---------------------------------------------------------------------------
// Live language switching
//
// The Python reference applies the language on startup only: demo.py loads the
// translators from cfg.language before the MainWindow is built, so the language
// takes effect on the next launch. The Go port switches live: the settings
// language ComboBox calls SwitchLanguage, which reloads and re-installs the
// ".qm" files, persists the Language config item and emits languageChanged so
// the application can rebuild its (now translated) UI.

var (
	languageMu       sync.Mutex
	languageChanged  []func()
	installedFluent  *gcommon.FluentTranslator
	installedGallery *GalleryTranslator
)

// OnLanguageChanged registers a listener for languageChanged.
func OnLanguageChanged(f func()) {
	languageMu.Lock()
	languageChanged = append(languageChanged, f)
	languageMu.Unlock()
}

// EmitLanguageChanged emits languageChanged.
func EmitLanguageChanged() {
	for _, f := range languageChangedSnapshot() {
		f()
	}
}

// InstallTranslators removes the currently installed fluent/gallery translators
// and installs fresh ones for the given locale.
func InstallTranslators(locale *qt.QLocale) {
	if installedFluent != nil {
		qt.QCoreApplication_RemoveTranslator(installedFluent.QTranslator)
		installedFluent.QTranslator.Delete()
		installedFluent = nil
	}
	if installedGallery != nil {
		qt.QCoreApplication_RemoveTranslator(installedGallery.QTranslator)
		installedGallery.QTranslator.Delete()
		installedGallery = nil
	}

	installedFluent = gcommon.NewFluentTranslator(locale, nil)
	installedGallery = NewGalleryTranslator(locale, nil)
	qt.QCoreApplication_InstallTranslator(installedFluent.QTranslator)
	qt.QCoreApplication_InstallTranslator(installedGallery.QTranslator)
}

// SwitchLanguage switches the UI language live: it reloads the ".qm" files,
// re-installs the translators, persists the Language config item and emits
// languageChanged so the application retranslates its widgets.
func SwitchLanguage(lang Language) {
	locale := lang.Locale()
	if lang != LanguageAuto {
		// Explicit locales are caller-owned (see Language.Locale); AUTO returns
		// the GoGC-armed system locale which must not be deleted.
		defer locale.Delete()
	}

	InstallTranslators(locale)
	gcommon.QConfigInstance.Set(ConfigInstance.Language, lang, true)
	EmitLanguageChanged()
}

func languageChangedSnapshot() []func() {
	languageMu.Lock()
	defer languageMu.Unlock()
	return append([]func(){}, languageChanged...)
}
