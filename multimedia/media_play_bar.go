package multimedia

import (
	"fmt"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/components/widgets"
	qt "github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/multimedia"
)

// MediaPlayBarButton is a transparent tool button sized for a media play bar.
type MediaPlayBarButton struct {
	*widgets.TransparentToolButton
}

// NewMediaPlayBarButton builds a play-bar button with the given icon.
func NewMediaPlayBarButton(icon interface{}, parent *qt.QWidget) *MediaPlayBarButton {
	b := &MediaPlayBarButton{TransparentToolButton: widgets.NewTransparentToolButtonIcon(icon, parent)}
	widgets.NewToolTipFilter(b.QWidget, 1000, widgets.ToolTipPositionTop)
	b.SetFixedSize2(30, 30)
	b.SetIconSize(qt.NewQSize2(16, 16))
	return b
}

// PlayButton toggles between the play and pause icons.
type PlayButton struct{ *MediaPlayBarButton }

// NewPlayButton builds a play button.
func NewPlayButton(parent *qt.QWidget) *PlayButton {
	b := &PlayButton{MediaPlayBarButton: NewMediaPlayBarButton(nil, parent)}
	b.SetIconSize(qt.NewQSize2(14, 14))
	b.SetPlay(false)
	return b
}

// SetPlay switches between the play and pause icons/tooltips.
func (b *PlayButton) SetPlay(isPlay bool) {
	if isPlay {
		b.SetIcon(common.PauseBold)
		b.SetToolTip("Pause")
	} else {
		b.SetIcon(common.PlaySolid)
		b.SetToolTip("Play")
	}
}

// paintRoundedBackground draws the shared rounded play-bar background.
func paintRoundedBackground(w *qt.QWidget) {
	painter := qt.NewQPainter2(w.QPaintDevice)
	defer painter.Delete()
	painter.SetRenderHints(qt.QPainter__Antialiasing)

	var bg, border *qt.QColor
	if common.IsDarkTheme() {
		bg = qt.NewQColor3(46, 46, 46)
		border = qt.NewQColor11(0, 0, 0, 20)
	} else {
		bg = qt.NewQColor3(248, 248, 248)
		border = qt.NewQColor11(0, 0, 0, 10)
	}
	defer bg.Delete()
	defer border.Delete()

	brush := qt.NewQBrush3(bg)
	painter.SetBrush(brush)
	brush.Delete()
	painter.SetPen(border)

	r := w.Rect()
	rect := r.Adjusted(1, 1, -1, -1) // GoGC-armed — do NOT Delete

	painter.DrawRoundedRect3(rect, 8, 8)
	painter.End()
}

// VolumeView is the flyout view hosting the mute button, volume slider and
// volume label. Its paint event overrides FlyoutViewBase's to use the darker
// play-bar palette.
type VolumeView struct {
	*widgets.FlyoutViewBase
	muteButton   *MediaPlayBarButton
	volumeSlider *widgets.Slider
	volumeLabel  *widgets.CaptionLabel
}

// NewVolumeView builds a volume flyout view.
func NewVolumeView(parent *qt.QWidget) *VolumeView {
	v := &VolumeView{FlyoutViewBase: widgets.NewFlyoutViewBase(parent)}
	v.muteButton = NewMediaPlayBarButton(common.Volume, v.QWidget)
	v.volumeSlider = widgets.NewSliderOrientation(qt.Horizontal, v.QWidget)
	v.volumeLabel = widgets.NewCaptionLabelText("30", v.QWidget)

	v.volumeSlider.SetRange(0, 100)
	v.volumeSlider.SetFixedWidth(208)
	v.SetFixedSize2(295, 64)

	h := v.Height()
	v.muteButton.Move(10, h/2-v.muteButton.Height()/2)
	v.volumeSlider.Move(45, 21)
	v.volumeLabel.AdjustSize()
	v.positionVolumeLabel()

	// Replace FlyoutViewBase's background paint with the play-bar palette.
	v.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		paintRoundedBackground(v.QWidget)
	})
	return v
}

// SetMuted updates the mute button state.
func (v *VolumeView) SetMuted(isMute bool) {
	if isMute {
		v.muteButton.SetIcon(common.Mute)
		v.muteButton.SetToolTip("Unmute")
	} else {
		v.muteButton.SetIcon(common.Volume)
		v.muteButton.SetToolTip("Mute")
	}
}

// SetVolume updates the slider and label.
func (v *VolumeView) SetVolume(volume int) {
	v.volumeSlider.SetValue(volume)
	v.volumeLabel.SetNum(volume)
	v.volumeLabel.AdjustSize()
	v.positionVolumeLabel()
}

// positionVolumeLabel moves the volume label so it sits immediately to the
// right of the volume slider, vertically centered (port of VolumeView.setVolume).
func (v *VolumeView) positionVolumeLabel() {
	tr := v.volumeLabel.FontMetrics().BoundingRectWithText(v.volumeLabel.Text()) // GoGC-armed — do NOT Delete
	v.volumeLabel.Move(v.Width()-20-tr.Width(), v.Height()/2-tr.Height()/2)
}

// VolumeButton is a play-bar button that shows a volume flyout.
type VolumeButton struct {
	*MediaPlayBarButton
	isMuted         bool
	volumeView      *VolumeView
	volumeFlyout    *widgets.Flyout
	onVolumeChanged func(volume int)
	onMutedChanged  func(muted bool)
}

// NewVolumeButton builds a volume button.
func NewVolumeButton(parent *qt.QWidget) *VolumeButton {
	b := &VolumeButton{MediaPlayBarButton: NewMediaPlayBarButton(nil, parent)}
	b.volumeView = NewVolumeView(b.QWidget)
	b.volumeFlyout = widgets.NewFlyout(b.volumeView.FlyoutViewBase, b.Window(), false)
	b.SetMuted(false)
	b.volumeFlyout.Hide()

	b.volumeView.muteButton.OnClicked(func() { b.emitMutedChanged(!b.isMuted) })
	b.volumeView.volumeSlider.OnValueChanged(func(volume int) { b.emitVolumeChanged(volume) })
	b.OnClicked(b.showVolumeFlyout)
	return b
}

// OnVolumeChanged registers the volumeChanged listener.
func (b *VolumeButton) OnVolumeChanged(f func(volume int)) { b.onVolumeChanged = f }

// OnMutedChanged registers the mutedChanged listener.
func (b *VolumeButton) OnMutedChanged(f func(muted bool)) { b.onMutedChanged = f }

func (b *VolumeButton) emitVolumeChanged(volume int) {
	if b.onVolumeChanged != nil {
		b.onVolumeChanged(volume)
	}
}

func (b *VolumeButton) emitMutedChanged(muted bool) {
	if b.onMutedChanged != nil {
		b.onMutedChanged(muted)
	}
}

// SetMuted updates the mute icon and the flyout view.
func (b *VolumeButton) SetMuted(isMute bool) {
	b.isMuted = isMute
	b.volumeView.SetMuted(isMute)
	if isMute {
		b.SetIcon(common.Mute)
	} else {
		b.SetIcon(common.Volume)
	}
}

// SetVolume updates the flyout view.
func (b *VolumeButton) SetVolume(volume int) { b.volumeView.SetVolume(volume) }

func (b *VolumeButton) showVolumeFlyout() {
	if b.volumeFlyout.IsVisible() {
		return
	}
	pos := pullUpFlyoutPosition(b.QWidget, b.volumeFlyout)
	b.volumeFlyout.Exec(pos, widgets.FlyoutAnimationPullUp)
	pos.Delete()
}

// pullUpFlyoutPosition mirrors qfluentwidgets PullUpFlyoutAnimationManager.
func pullUpFlyoutPosition(target *qt.QWidget, flyout *widgets.Flyout) *qt.QPoint {
	sh := flyout.SizeHint()
	fw := sh.Width()
	fh := sh.Height()
	m := flyout.Layout().ContentsMargins()

	origin := qt.NewQPoint2(0, 0)
	pos := target.MapToGlobal(origin) // GoGC-armed — do NOT Delete
	origin.Delete()

	x := pos.X() + target.Width()/2 - fw/2
	y := pos.Y() - fh + m.Bottom()
	return qt.NewQPoint2(x, y)
}

// MediaPlayBarBase is the shared base of the media play bars.
type MediaPlayBarBase struct {
	*qt.QWidget
	Player                MediaPlayerBase
	playButton            *PlayButton
	volumeButton          *VolumeButton
	progressSlider        *widgets.Slider
	opacityEffect         *qt.QGraphicsOpacityEffect
	opacityAni            *qt.QPropertyAnimation
	onPositionChangedHook func(position int64)
}

func newMediaPlayBarBase(parent *qt.QWidget) *MediaPlayBarBase {
	w := &MediaPlayBarBase{QWidget: qt.NewQWidget(parent)}
	w.playButton = NewPlayButton(w.QWidget)
	w.volumeButton = NewVolumeButton(w.QWidget)
	w.progressSlider = widgets.NewSliderOrientation(qt.Horizontal, w.QWidget)

	w.opacityEffect = qt.NewQGraphicsOpacityEffect2(w.QObject)
	w.opacityAni = qt.NewQPropertyAnimation2(w.opacityEffect.QObject, []byte("opacity"))
	w.opacityEffect.SetOpacity(1)
	w.opacityAni.SetDuration(250)
	w.SetGraphicsEffect(w.opacityEffect.QGraphicsEffect)

	common.FluentStyleSheet(common.FluentMediaPlayer).Apply(w.QWidget, common.ThemeAuto)
	w.playButton.OnClicked(w.TogglePlayState)

	w.OnPaintEvent(func(super func(e *qt.QPaintEvent), e *qt.QPaintEvent) {
		paintRoundedBackground(w.QWidget)
	})
	return w
}

// SetMediaPlayer binds a media player to the play bar.
func (w *MediaPlayBarBase) SetMediaPlayer(player MediaPlayerBase) {
	w.Player = player

	player.OnDurationChanged(func(duration int64) { w.progressSlider.SetMaximum(int(duration)) })
	player.OnPositionChanged(func(position int64) {
		w.progressSlider.SetValue(int(position))
		if w.onPositionChangedHook != nil {
			w.onPositionChangedHook(position)
		}
	})
	player.OnMediaStatusChanged(func(status multimedia.QMediaPlayer__MediaStatus) {
		w.playButton.SetPlay(player.IsPlaying())
	})
	player.OnVolumeChanged(func(volume int) { w.volumeButton.SetVolume(volume) })
	player.OnMutedChanged(func(muted bool) { w.volumeButton.SetMuted(muted) })

	w.progressSlider.OnSliderMoved(func(position int) { player.SetPosition(int64(position)) })
	w.progressSlider.OnClicked(func(position int) { player.SetPosition(int64(position)) })
	w.volumeButton.OnVolumeChanged(func(volume int) { player.SetVolume(volume) })
	w.volumeButton.OnMutedChanged(func(muted bool) { player.SetMuted(muted) })

	player.SetVolume(30)
}

// FadeIn fades the play bar in.
func (w *MediaPlayBarBase) FadeIn() {
	w.opacityAni.SetStartValue(qt.NewQVariant12(w.opacityEffect.Opacity()))
	w.opacityAni.SetEndValue(qt.NewQVariant12(1))
	w.opacityAni.Start()
}

// FadeOut fades the play bar out.
func (w *MediaPlayBarBase) FadeOut() {
	w.opacityAni.SetStartValue(qt.NewQVariant12(w.opacityEffect.Opacity()))
	w.opacityAni.SetEndValue(qt.NewQVariant12(0))
	w.opacityAni.Start()
}

// Play starts or resumes playback.
func (w *MediaPlayBarBase) Play() {
	if w.Player != nil {
		w.Player.Play()
	}
}

// Pause pauses playback.
func (w *MediaPlayBarBase) Pause() {
	if w.Player != nil {
		w.Player.Pause()
	}
}

// Stop stops playback.
func (w *MediaPlayBarBase) Stop() {
	if w.Player != nil {
		w.Player.Stop()
	}
}

// SetVolume sets the player volume.
func (w *MediaPlayBarBase) SetVolume(volume int) {
	if w.Player != nil {
		w.Player.SetVolume(volume)
	}
}

// SetPosition sets the playback position in milliseconds.
func (w *MediaPlayBarBase) SetPosition(position int) {
	if w.Player != nil {
		w.Player.SetPosition(int64(position))
	}
}

// TogglePlayState toggles between play and pause.
func (w *MediaPlayBarBase) TogglePlayState() {
	if w.Player == nil {
		return
	}
	if w.Player.IsPlaying() {
		w.Player.Pause()
	} else {
		w.Player.Play()
	}
	w.playButton.SetPlay(w.Player.IsPlaying())
}

// SimpleMediaPlayBar is a single-row play bar.
type SimpleMediaPlayBar struct {
	*MediaPlayBarBase
	hBoxLayout *qt.QHBoxLayout
}

// NewSimpleMediaPlayBar builds a simple play bar.
func NewSimpleMediaPlayBar(parent *qt.QWidget) *SimpleMediaPlayBar {
	w := &SimpleMediaPlayBar{MediaPlayBarBase: newMediaPlayBarBase(parent)}
	w.hBoxLayout = qt.NewQHBoxLayout(w.QWidget)
	w.hBoxLayout.SetContentsMargins(10, 4, 10, 4)
	w.hBoxLayout.SetSpacing(6)
	w.hBoxLayout.AddWidget3(w.playButton.QWidget, 0, qt.AlignLeft)
	w.hBoxLayout.AddWidget2(w.progressSlider.QWidget, 1)
	w.hBoxLayout.AddWidget(w.volumeButton.QWidget)

	w.SetFixedHeight(48)
	w.SetMediaPlayer(NewMediaPlayer(w.QObject))
	return w
}

// AddButton adds a button to the right side of the play bar.
func (w *SimpleMediaPlayBar) AddButton(button *MediaPlayBarButton) {
	w.hBoxLayout.AddWidget(button.QWidget)
}

// StandardMediaPlayBar is the two-row standard play bar.
type StandardMediaPlayBar struct {
	*MediaPlayBarBase
	vBoxLayout            *qt.QVBoxLayout
	timeLayout            *qt.QHBoxLayout
	buttonLayout          *qt.QHBoxLayout
	leftButtonContainer   *qt.QWidget
	centerButtonContainer *qt.QWidget
	rightButtonContainer  *qt.QWidget
	leftButtonLayout      *qt.QHBoxLayout
	centerButtonLayout    *qt.QHBoxLayout
	rightButtonLayout     *qt.QHBoxLayout
	skipBackButton        *MediaPlayBarButton
	skipForwardButton     *MediaPlayBarButton
	currentTimeLabel      *widgets.CaptionLabel
	remainTimeLabel       *widgets.CaptionLabel
}

// NewStandardMediaPlayBar builds a standard play bar.
func NewStandardMediaPlayBar(parent *qt.QWidget) *StandardMediaPlayBar {
	w := &StandardMediaPlayBar{MediaPlayBarBase: newMediaPlayBarBase(parent)}
	w.vBoxLayout = qt.NewQVBoxLayout(w.QWidget)
	w.timeLayout = qt.NewQHBoxLayout2()
	w.buttonLayout = qt.NewQHBoxLayout2()
	w.leftButtonContainer = qt.NewQWidget2()
	w.centerButtonContainer = qt.NewQWidget2()
	w.rightButtonContainer = qt.NewQWidget2()
	w.leftButtonLayout = qt.NewQHBoxLayout(w.leftButtonContainer)
	w.centerButtonLayout = qt.NewQHBoxLayout(w.centerButtonContainer)
	w.rightButtonLayout = qt.NewQHBoxLayout(w.rightButtonContainer)

	w.skipBackButton = NewMediaPlayBarButton(common.SkipBack, w.QWidget)
	w.skipForwardButton = NewMediaPlayBarButton(common.SkipForward, w.QWidget)
	w.currentTimeLabel = widgets.NewCaptionLabelText("0:00:00", w.QWidget)
	w.remainTimeLabel = widgets.NewCaptionLabelText("0:00:00", w.QWidget)

	w.onPositionChangedHook = w.updateTimeLabels
	w.initWidgets()
	return w
}

func (w *StandardMediaPlayBar) initWidgets() {
	w.SetFixedHeight(102)
	w.vBoxLayout.SetSpacing(6)
	w.vBoxLayout.SetContentsMargins(5, 9, 5, 9)
	w.vBoxLayout.AddWidget3(w.progressSlider.QWidget, 1, qt.AlignTop)

	w.vBoxLayout.AddLayout(w.timeLayout.QLayout)
	w.timeLayout.SetContentsMargins(10, 0, 10, 0)
	w.timeLayout.AddWidget3(w.currentTimeLabel.QWidget, 0, qt.AlignLeft)
	w.timeLayout.AddWidget3(w.remainTimeLabel.QWidget, 0, qt.AlignRight)

	w.vBoxLayout.AddStretchWithStretch(1)
	w.vBoxLayout.AddLayout2(w.buttonLayout.QLayout, 1)
	w.buttonLayout.SetContentsMargins(0, 0, 0, 0)
	w.leftButtonLayout.SetContentsMargins(4, 0, 0, 0)
	w.centerButtonLayout.SetContentsMargins(0, 0, 0, 0)
	w.rightButtonLayout.SetContentsMargins(0, 0, 4, 0)

	w.leftButtonLayout.AddWidget3(w.volumeButton.QWidget, 0, qt.AlignLeft)
	w.centerButtonLayout.AddWidget(w.skipBackButton.QWidget)
	w.centerButtonLayout.AddWidget(w.playButton.QWidget)
	w.centerButtonLayout.AddWidget(w.skipForwardButton.QWidget)

	w.buttonLayout.AddWidget3(w.leftButtonContainer, 0, qt.AlignLeft)
	w.buttonLayout.AddWidget3(w.centerButtonContainer, 0, qt.AlignHCenter)
	w.buttonLayout.AddWidget3(w.rightButtonContainer, 0, qt.AlignRight)

	w.SetMediaPlayer(NewMediaPlayer(w.QObject))

	w.skipBackButton.OnClicked(func() { w.SkipBack(10000) })
	w.skipForwardButton.OnClicked(func() { w.SkipForward(30000) })
}

// SkipBack rewinds playback by ms milliseconds.
func (w *StandardMediaPlayBar) SkipBack(ms int) {
	w.Player.SetPosition(w.Player.Position() - int64(ms))
}

// SkipForward fast-forwards playback by ms milliseconds.
func (w *StandardMediaPlayBar) SkipForward(ms int) {
	w.Player.SetPosition(w.Player.Position() + int64(ms))
}

func (w *StandardMediaPlayBar) updateTimeLabels(position int64) {
	w.currentTimeLabel.SetText(w.formatTime(position))
	w.remainTimeLabel.SetText(w.formatTime(w.Player.Duration() - position))
}

func (w *StandardMediaPlayBar) formatTime(ms int64) string {
	t := int(ms / 1000)
	s := t % 60
	m := (t / 60) % 60
	h := t / 3600
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}
