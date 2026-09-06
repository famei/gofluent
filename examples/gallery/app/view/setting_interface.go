package view

import (
	"strconv"

	gcommon "github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/layout"
	"github.com/famei/gofluent/components/settings"
	"github.com/famei/gofluent/components/widgets"
	gallerycommon "github.com/famei/gofluent/examples/gallery/app/common"
	qt "github.com/mappu/miqt/qt"
)

// SettingInterface is the "Settings" page (port of app/view/setting_interface.py).
type SettingInterface struct {
	*widgets.ScrollArea
	scrollWidget *qt.QWidget
	expandLayout *layout.ExpandLayout
	settingLabel *qt.QLabel

	musicInThisPCGroup *settings.SettingCardGroup
	musicFolderCard    *settings.FolderListSettingCard
	downloadFolderCard *settings.PushSettingCard

	personalGroup  *settings.SettingCardGroup
	micaCard       *settings.SwitchSettingCard
	themeCard      *settings.OptionsSettingCard
	themeColorCard *settings.CustomColorSettingCard
	zoomCard       *settings.OptionsSettingCard
	languageCard   *settings.ComboBoxSettingCard

	materialGroup  *settings.SettingCardGroup
	blurRadiusCard *settings.RangeSettingCard

	updateSoftwareGroup *settings.SettingCardGroup
	updateOnStartUpCard *settings.SwitchSettingCard

	aboutGroup   *settings.SettingCardGroup
	helpCard     *settings.HyperlinkCard
	feedbackCard *settings.PrimaryPushSettingCard
	aboutCard    *settings.PrimaryPushSettingCard
}

// NewSettingInterface builds the setting interface.
func NewSettingInterface(parent *qt.QWidget) *SettingInterface {
	cfg := gallerycommon.ConfigInstance
	i := &SettingInterface{ScrollArea: widgets.NewScrollArea(parent)}
	i.scrollWidget = qt.NewQWidget2()
	i.expandLayout = layout.NewExpandLayout(i.scrollWidget)

	i.settingLabel = qt.NewQLabel5(gallerycommon.Tr("SettingInterface", "Settings"), i.QWidget)

	i.musicInThisPCGroup = settings.NewSettingCardGroup(gallerycommon.Tr("SettingInterface", "Music on this PC"), i.scrollWidget)
	i.musicFolderCard = settings.NewFolderListSettingCard(
		cfg.MusicFolders, gallerycommon.Tr("SettingInterface", "Local music library"), "", "", i.musicInThisPCGroup.QWidget)
	i.downloadFolderCard = settings.NewPushSettingCard(
		gallerycommon.Tr("SettingInterface", "Choose folder"), gcommon.Download,
		gallerycommon.Tr("SettingInterface", "Download directory"),
		cfg.Get(cfg.DownloadFolder).(string), i.musicInThisPCGroup.QWidget)

	i.personalGroup = settings.NewSettingCardGroup(gallerycommon.Tr("SettingInterface", "Personalization"), i.scrollWidget)
	i.micaCard = settings.NewSwitchSettingCard(
		gcommon.Transparent, gallerycommon.Tr("SettingInterface", "Mica effect"),
		gallerycommon.Tr("SettingInterface", "Apply semi transparent to windows and surfaces"),
		cfg.MicaEnabled, i.personalGroup.QWidget)
	i.themeCard = settings.NewOptionsSettingCard(
		cfg.ThemeMode(), gcommon.Brush,
		gallerycommon.Tr("SettingInterface", "Application theme"),
		gallerycommon.Tr("SettingInterface", "Change the appearance of your application"),
		[]string{gallerycommon.Tr("SettingInterface", "Light"), gallerycommon.Tr("SettingInterface", "Dark"),
			gallerycommon.Tr("SettingInterface", "Use system setting")}, i.personalGroup.QWidget)
	i.themeColorCard = settings.NewCustomColorSettingCard(
		cfg.ThemeColor(), gcommon.Palette, gallerycommon.Tr("SettingInterface", "Theme color"),
		gallerycommon.Tr("SettingInterface", "Change the theme color of you application"),
		i.personalGroup.QWidget, false)
	i.zoomCard = settings.NewOptionsSettingCard(
		cfg.DpiScale, gcommon.Zoom, gallerycommon.Tr("SettingInterface", "Interface zoom"),
		gallerycommon.Tr("SettingInterface", "Change the size of widgets and fonts"),
		[]string{"100%", "125%", "150%", "175%", "200%",
			gallerycommon.Tr("SettingInterface", "Use system setting")}, i.personalGroup.QWidget)
	i.languageCard = settings.NewComboBoxSettingCard(
		cfg.Language, gcommon.Language, gallerycommon.Tr("SettingInterface", "Language"),
		gallerycommon.Tr("SettingInterface", "Set your preferred language for UI"),
		[]string{"简体中文", "繁體中文", "English",
			gallerycommon.Tr("SettingInterface", "Use system setting")}, i.personalGroup.QWidget)

	i.materialGroup = settings.NewSettingCardGroup(gallerycommon.Tr("SettingInterface", "Material"), i.scrollWidget)
	i.blurRadiusCard = settings.NewRangeSettingCard(
		cfg.BlurRadius, gcommon.Album, gallerycommon.Tr("SettingInterface", "Acrylic blur radius"),
		gallerycommon.Tr("SettingInterface", "The greater the radius, the more blurred the image"), i.materialGroup.QWidget)

	i.updateSoftwareGroup = settings.NewSettingCardGroup(gallerycommon.Tr("SettingInterface", "Software update"), i.scrollWidget)
	i.updateOnStartUpCard = settings.NewSwitchSettingCard(
		gcommon.Update, gallerycommon.Tr("SettingInterface", "Check for updates when the application starts"),
		gallerycommon.Tr("SettingInterface", "The new version will be more stable and have more features"),
		cfg.CheckUpdateAtStartUp, i.updateSoftwareGroup.QWidget)

	i.aboutGroup = settings.NewSettingCardGroup(gallerycommon.Tr("SettingInterface", "About"), i.scrollWidget)
	i.helpCard = settings.NewHyperlinkCard(
		gallerycommon.HELP_URL, gallerycommon.Tr("SettingInterface", "Open help page"),
		gcommon.Help, gallerycommon.Tr("SettingInterface", "Help"),
		gallerycommon.Tr("SettingInterface", "Discover new features and learn useful tips about PyQt-Fluent-Widgets"),
		i.aboutGroup.QWidget)
	i.feedbackCard = settings.NewPrimaryPushSettingCard(
		gallerycommon.Tr("SettingInterface", "Provide feedback"), gcommon.Feedback,
		gallerycommon.Tr("SettingInterface", "Provide feedback"),
		gallerycommon.Tr("SettingInterface", "Help us improve PyQt-Fluent-Widgets by providing feedback"),
		i.aboutGroup.QWidget)
	aboutContent := "© " + gallerycommon.Tr("SettingInterface", "Copyright") + " " + strconv.Itoa(gallerycommon.YEAR) + ", " + gallerycommon.AUTHOR + ". " +
		gallerycommon.Tr("SettingInterface", "Version") + " " + gallerycommon.VERSION
	i.aboutCard = settings.NewPrimaryPushSettingCard(
		gallerycommon.Tr("SettingInterface", "Check update"), gcommon.Info,
		gallerycommon.Tr("SettingInterface", "About"), aboutContent, i.aboutGroup.QWidget)

	i.initWidget()
	return i
}

func (i *SettingInterface) initWidget() {
	i.Resize(1000, 800)
	i.SetHorizontalScrollBarPolicy(qt.ScrollBarAlwaysOff)
	i.SetViewportMargins(0, 80, 0, 20)
	i.SetWidget(i.scrollWidget)
	i.SetWidgetResizable(true)
	i.SetObjectName("settingInterface")

	i.scrollWidget.SetObjectName("scrollWidget")
	i.settingLabel.SetObjectName("settingLabel")
	gallerycommon.StyleSettingInterface.Apply(i.QWidget, gcommon.ThemeAuto)

	i.micaCard.SetEnabled(gallerycommon.IsWin11())

	i.initLayout()
	i.connectSignalToSlot()
}

func (i *SettingInterface) initLayout() {
	i.settingLabel.Move(36, 30)

	i.musicInThisPCGroup.AddSettingCard(i.musicFolderCard.QWidget)
	i.musicInThisPCGroup.AddSettingCard(i.downloadFolderCard.QWidget)

	i.personalGroup.AddSettingCard(i.micaCard.QWidget)
	i.personalGroup.AddSettingCard(i.themeCard.QWidget)
	i.personalGroup.AddSettingCard(i.themeColorCard.QWidget)
	i.personalGroup.AddSettingCard(i.zoomCard.QWidget)
	i.personalGroup.AddSettingCard(i.languageCard.QWidget)

	i.materialGroup.AddSettingCard(i.blurRadiusCard.QWidget)

	i.updateSoftwareGroup.AddSettingCard(i.updateOnStartUpCard.QWidget)

	i.aboutGroup.AddSettingCard(i.helpCard.QWidget)
	i.aboutGroup.AddSettingCard(i.feedbackCard.QWidget)
	i.aboutGroup.AddSettingCard(i.aboutCard.QWidget)

	i.expandLayout.SetSpacing(28)
	i.expandLayout.SetContentsMargins(36, 10, 36, 0)
	i.expandLayout.AddWidget(i.musicInThisPCGroup.QWidget)
	i.expandLayout.AddWidget(i.personalGroup.QWidget)
	i.expandLayout.AddWidget(i.materialGroup.QWidget)
	i.expandLayout.AddWidget(i.updateSoftwareGroup.QWidget)
	i.expandLayout.AddWidget(i.aboutGroup.QWidget)
}

func (i *SettingInterface) showRestartTooltip() {
	widgets.InfoBarSuccess(
		gallerycommon.Tr("SettingInterface", "Updated successfully"),
		gallerycommon.Tr("SettingInterface", "Configuration takes effect after restart"), i.QWidget)
}

func (i *SettingInterface) onDownloadFolderCardClicked() {
	folder := qt.QFileDialog_GetExistingDirectory3(i.QWidget, gallerycommon.Tr("SettingInterface", "Choose folder"), "./")
	if folder == "" || gallerycommon.ConfigInstance.Get(gallerycommon.ConfigInstance.DownloadFolder) == folder {
		return
	}
	gallerycommon.ConfigInstance.Set(gallerycommon.ConfigInstance.DownloadFolder, folder, false)
	i.downloadFolderCard.SetContent(folder)
}

func (i *SettingInterface) connectSignalToSlot() {
	gcommon.QConfigInstance.OnAppRestart(i.showRestartTooltip)

	i.downloadFolderCard.OnClicked = i.onDownloadFolderCardClicked

	gcommon.QConfigInstance.OnThemeChanged(func(t gcommon.Theme) { gcommon.SetTheme(t, false, true) })
	i.micaCard.OnCheckedChanged = func(isChecked bool) { gallerycommon.SignalBusInstance.EmitMicaEnableChanged(isChecked) }

	// Language: switch the UI language live. SwitchLanguage reloads the .qm
	// files, persists the Language config item and emits languageChanged, which
	// rebuilds the main window against the freshly installed translators.
	i.languageCard.ComboBox().OnCurrentIndexChanged = func(index int) {
		lang, ok := i.languageCard.ComboBox().ItemData(index).(gallerycommon.Language)
		if !ok {
			return
		}
		gallerycommon.SwitchLanguage(lang)
	}

	i.feedbackCard.OnClicked = func() { qt.QDesktopServices_OpenUrl(qt.NewQUrl3(gallerycommon.FEEDBACK_URL)) }
}
