#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "$0")"

export PATH=/usr/lib/mxe/usr/bin:$PATH
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:${GOPATH:-$HOME/go}/bin
export GOPROXY=https://goproxy.io,direct

export CXX=x86_64-w64-mingw32.static-g++
export CC=x86_64-w64-mingw32.static-gcc
export PKG_CONFIG=x86_64-w64-mingw32.static-pkg-config
export PKG_CONFIG_PATH=/usr/lib/mxe/usr/x86_64-w64-mingw32.static/lib/pkgconfig:/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/lib/pkgconfig

export CGO_LDFLAGS='-L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/platforms -lqwindows -lQt5FontDatabaseSupport -lQt5EventDispatcherSupport -lQt5ThemeSupport -lQt5PlatformCompositorSupport -lQt5AccessibilitySupport -lQt5WindowsUIAutomationSupport -lwtsapi32 -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/styles -lqwindowsvistastyle -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/lib -ljpeg -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/imageformats -lqjpeg -lqico -lqgif'

export GOOS=windows
export CGO_ENABLED=1
export GOFLAGS=-buildvcs=false
export CGO_CXXFLAGS="-std=c++11"

go build -ldflags "-s -w" --tags=windowsqtstatic -o gallery.exe ./examples/gallery