package multimedia

import (
	qt "github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/multimedia"
)

// MediaPlayerBase is the interface implemented by media players consumed by the
// play bars and video widget. It replaces the abstract Python MediaPlayerBase
// class (whose pyqtSignal declarations are represented by the On* registrations
// here).
type MediaPlayerBase interface {
	IsPlaying() bool
	Duration() int64
	Position() int64
	Volume() int
	IsMuted() bool
	SetPosition(position int64)
	SetVolume(volume int)
	SetMuted(muted bool)
	SetSource(url *qt.QUrl)
	SetVideoOutput(output *multimedia.QGraphicsVideoItem)
	Play()
	Pause()
	Stop()
	OnDurationChanged(func(duration int64))
	OnPositionChanged(func(position int64))
	OnMediaStatusChanged(func(status multimedia.QMediaPlayer__MediaStatus))
	OnVolumeChanged(func(volume int))
	OnMutedChanged(func(muted bool))
}

// MediaPlayer wraps QtMultimedia's QMediaPlayer and adds the fluent convenience
// API (IsPlaying, Source/SetSource, PlaybackState, SetVideoOutput).
type MediaPlayer struct {
	*multimedia.QMediaPlayer
}

// NewMediaPlayer builds a media player.
func NewMediaPlayer(parent *qt.QObject) *MediaPlayer {
	var p *MediaPlayer
	if parent != nil {
		p = &MediaPlayer{QMediaPlayer: multimedia.NewQMediaPlayer2(parent)}
	} else {
		p = &MediaPlayer{QMediaPlayer: multimedia.NewQMediaPlayer()}
	}

	// Compute (and discard) the canonical URL whenever the media changes, as the
	// Python port does in MediaPlayer.__init__.
	p.OnMediaChanged(func(media *multimedia.QMediaContent) {
		if media == nil {
			return
		}
		_ = media.CanonicalUrl() // GoGC-armed — do NOT Delete
	})
	p.SetNotifyInterval(1000)
	return p
}

// IsPlaying reports whether the media is playing.
func (p *MediaPlayer) IsPlaying() bool {
	return p.State() == multimedia.QMediaPlayer__PlayingState
}

// Source returns the canonical URL of the current media.
func (p *MediaPlayer) Source() *qt.QUrl {
	m := p.CurrentMedia() // GoGC-armed — do NOT Delete
	return m.CanonicalUrl()
}

// SetSource sets the current media source from a URL.
func (p *MediaPlayer) SetSource(url *qt.QUrl) {
	m := multimedia.NewQMediaContent2(url)
	p.SetMedia(m)
	m.Delete()
}

// PlaybackState returns the playback state (Python playbackState parity).
func (p *MediaPlayer) PlaybackState() multimedia.QMediaPlayer__State { return p.State() }

// MediaStatus returns the current media status (Python mediaStatus parity).
func (p *MediaPlayer) MediaStatus() multimedia.QMediaPlayer__MediaStatus {
	return p.QMediaPlayer.MediaStatus()
}

// SetVideoOutput binds a graphics video item as the video sink.
func (p *MediaPlayer) SetVideoOutput(output *multimedia.QGraphicsVideoItem) {
	p.QMediaPlayer.SetVideoOutputWithVideoOutput(output)
}
