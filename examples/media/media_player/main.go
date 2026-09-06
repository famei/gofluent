// Command media_player migrates examples/media/media_player/demo.py.
//
// Demo1 hosts a SimpleMediaPlayBar and a StandardMediaPlayBar in a vertical
// layout: the simple bar streams a remote track, the standard bar plays a local
// mp3. Because a statically linked exe has no runtime resource path, the mp3 is
// embedded with go:embed and written to a temp file at startup so QUrl can
// resolve a real local file. Demo2 hosts a VideoWidget streaming a remote video.
package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/famei/gofluent/common"
	"github.com/famei/gofluent/examples/internal/demo"
	"github.com/famei/gofluent/multimedia"
	qt "github.com/mappu/miqt/qt"
)

//go:embed resource/*.mp3
var mediaFS embed.FS

func newDemo1() *qt.QWidget {
	common.SetTheme(common.ThemeDark, false, false)

	w := qt.NewQWidget2()
	w.Resize(500, 300)
	vBoxLayout := qt.NewQVBoxLayout(w)

	simplePlayBar := multimedia.NewSimpleMediaPlayBar(w)
	standardPlayBar := multimedia.NewStandardMediaPlayBar(w)
	vBoxLayout.AddWidget(simplePlayBar.QWidget)
	vBoxLayout.AddWidget(standardPlayBar.QWidget)

	// Online music.
	url := qt.NewQUrl3("https://files.cnblogs.com/files/blogs/677826/beat.zip?t=1693900324")
	simplePlayBar.Player.SetSource(url)
	url.Delete()

	// Local music: materialize the embedded mp3 and build a local-file URL.
	url = localFileURL()
	standardPlayBar.Player.SetSource(url)

	return w
}

// localFileURL writes the embedded mp3 to a temp file and returns a local-file
// URL for it. QMediaContent copies the QUrl, so the caller owns the result.
func localFileURL() *qt.QUrl {
	mp3, err := mediaFS.ReadFile("resource/aiko - シアワセ.mp3")
	if err != nil {
		panic(err)
	}
	tmp := filepath.Join(os.TempDir(), "aiko.mp3")
	if err := os.WriteFile(tmp, mp3, 0o600); err != nil {
		panic(err)
	}
	return qt.QUrl_FromLocalFile(tmp)
}

func newDemo2() *qt.QWidget {
	w := qt.NewQWidget2()
	vBoxLayout := qt.NewQVBoxLayout(w)
	videoWidget := multimedia.NewVideoWidget(w)

	url := qt.NewQUrl3("https://media.w3.org/2010/05/sintel/trailer.mp4")
	videoWidget.SetVideo(url)
	url.Delete()
	videoWidget.Play()

	vBoxLayout.SetContentsMargins(0, 0, 0, 0)
	vBoxLayout.AddWidget(videoWidget.QWidget)
	w.Resize(800, 450)
	return w
}

func main() {
	demo.Run(newDemo1, newDemo2)
}
