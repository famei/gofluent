#!/usr/bin/env bash
# build_windows.sh — cross-compile gofluent examples for Windows x86_64 (static Qt5).
#
# MUST run inside WSL Ubuntu 24.04 with the MXE cross toolchain installed at
# /usr/lib/mxe. The Windows host cannot run MXE directly.
#
# Two modes (env selectable):
#   * default       : `go vet ./...` compiles the whole module (library +
#                     examples) against static Qt5, then fully static-links the
#                     representative gallery app (and the settings window app)
#                     to .exe files when those packages exist yet.
#   * FULL_BUILD=1  : additionally static-links every example `main` package to
#                     its own .exe under dist/ (slow, but complete).
#
# Usage (from the gofluent module root):
#     cd /path/to/gofluent
#     bash examples/build_windows.sh                # fast verify
#     FULL_BUILD=1 bash examples/build_windows.sh   # every demo .exe

set -euo pipefail
cd "$(dirname "$0")/.."   # gofluent module root

# --- Toolchain / proxy (TASK.md §Constraints) ---------------------------------
export PATH=/usr/lib/mxe/usr/bin:$PATH
export GOROOT=/usr/local/go
export GOPATH=${GOPATH:-$HOME/go}
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
export GOPROXY=https://goproxy.io,direct
export GOPRIVATE=git.mycompany.com,github.com/my/private

# --- MXE cross toolchain ------------------------------------------------------
export CXX=x86_64-w64-mingw32.static-g++
export CC=x86_64-w64-mingw32.static-gcc
export PKG_CONFIG=x86_64-w64-mingw32.static-pkg-config
# PKG_CONFIG_PATH is required so the static pkg-config finds Qt5*.pc and the
# MXE static libs. NOTE: gofluent/build_windows.sh omitted this line; the
# examples script carries it so cross-compilation actually resolves Qt5Widgets.
export PKG_CONFIG_PATH=/usr/lib/mxe/usr/x86_64-w64-mingw32.static/lib/pkgconfig:/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/lib/pkgconfig

# --- Static Qt5 link flags (TASK.md §Constraints) -----------------------------
export CGO_LDFLAGS='-L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/platforms -lqwindows -lQt5FontDatabaseSupport -lQt5EventDispatcherSupport -lQt5ThemeSupport -lQt5PlatformCompositorSupport -lQt5AccessibilitySupport -lQt5WindowsUIAutomationSupport -lwtsapi32 -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/styles -lqwindowsvistastyle -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/lib -ljpeg -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/imageformats -lqjpeg -lqico'

# --- Go build settings --------------------------------------------------------
export GOOS=windows
export CGO_ENABLED=1
export GOFLAGS=-buildvcs=false
export CGO_CXXFLAGS="-std=c++11"

# build_exe <output> <package> — static-link one package to a Windows .exe.
build_exe() {
    local out="$1" pkg="$2"
    go build \
        -gcflags=-trimpath="${GOPATH}" \
        -asmflags=-trimpath="${GOPATH}" \
        -ldflags "-s -w" \
        --tags=windowsqtstatic \
        -o "$out" "$pkg"
}

echo "==> go version"
go version
echo "==> pkg-config Qt5Widgets"
$PKG_CONFIG --exists Qt5Widgets && echo "Qt5Widgets OK" || { echo "Qt5Widgets FAIL"; exit 1; }

echo "==> go vet ./... (compile-level check of library + examples, static Qt5)"
go vet --tags=windowsqtstatic ./...

mkdir -p dist

# Representative full static links: gallery (flagship app) and the settings
# window app (window app with .ui/.qrc resources). Guarded so the script still
# works while those packages are being implemented.
if [[ -f examples/gallery/main.go ]]; then
    echo "==> static-link representative app: gallery"
    build_exe dist/gallery.exe ./examples/gallery
fi
if [[ -f examples/window/settings/main.go ]]; then
    echo "==> static-link representative app: settings window"
    build_exe dist/window_settings.exe ./examples/window/settings
fi

if [[ "${FULL_BUILD:-0}" == "1" ]]; then
    echo "==> FULL_BUILD=1: static-linking every example main package"
    mapfile -t DEMOS < <(go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./examples/...)
    for pkg in "${DEMOS[@]}"; do
        [[ -z "$pkg" ]] && continue
        name="${pkg#gofluent/examples/}"
        exe="dist/${name//\//_}.exe"
        echo ">> $pkg -> $exe"
        build_exe "$exe" "$pkg"
    done
else
    echo "==> (FULL_BUILD unset; skipped per-demo .exe linking. Set FULL_BUILD=1 for complete build)"
fi

echo "==> DONE"
