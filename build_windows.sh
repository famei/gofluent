#!/usr/bin/env bash
# build_windows.sh — static cross-compile gofluent for Windows x86_64.
#
# MUST be executed inside WSL Ubuntu 24.04 with the MXE cross toolchain
# installed at /usr/lib/mxe. The Windows host cannot run MXE directly.
#
# gofluent is a library migration (PyQt-Fluent-Widgets -> Go/miqt): it has no
# `package main` yet, so this script verifies that every package cross-compiles
# for windows/amd64 against static Qt5 (the TASK.md verification goal). Once an
# application entry point (`package main`) is added, switch the final command to
# the commented-out `-o` form at the bottom to emit the .exe.
#
# Usage (from the gofluent module directory):
#     cd /path/to/gofluent
#     bash build_windows.sh
#
# The produced binary is a static Qt5 executable that also needs the runtime
# PATH documented in TASK.md (msys64 ucrt64/qt5-static plugin directories) if
# static plugin auto-discovery is ever disabled; the windowsqtstatic tag plus
# the CGO_LDFLAGS below link the platform/style/image plugins in.

set -euo pipefail

# --- Toolchain / proxy (TASK.md §Constraints) ---------------------------------
export PATH=/usr/lib/mxe/usr/bin:$PATH
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:${GOPATH:-$HOME/go}/bin
export GOPROXY=https://goproxy.io,direct
export GOPRIVATE=git.mycompany.com,github.com/my/private

# --- MXE cross toolchain ------------------------------------------------------
export CXX=x86_64-w64-mingw32.static-g++
export CC=x86_64-w64-mingw32.static-gcc
export PKG_CONFIG=x86_64-w64-mingw32.static-pkg-config
export PKG_CONFIG_PATH=/usr/lib/mxe/usr/x86_64-w64-mingw32.static/lib/pkgconfig:/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/lib/pkgconfig

# --- Static Qt5 link flags (TASK.md §Constraints) -----------------------------
export CGO_LDFLAGS='-L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/platforms -lqwindows -lQt5FontDatabaseSupport -lQt5EventDispatcherSupport -lQt5ThemeSupport -lQt5PlatformCompositorSupport -lQt5AccessibilitySupport -lQt5WindowsUIAutomationSupport -lwtsapi32 -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/styles -lqwindowsvistastyle -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/lib -ljpeg -L/usr/lib/mxe/usr/x86_64-w64-mingw32.static/qt5/plugins/imageformats -lqjpeg -lqico'

# --- Go build settings --------------------------------------------------------
export GOOS=windows
export CGO_ENABLED=1
export GOFLAGS=-buildvcs=false
export CGO_CXXFLAGS="-std=c++11"

echo "==> Cross-compiling gofluent packages for windows/amd64 (static Qt5 via MXE)"
echo "    CC=$CC"
echo "    CXX=$CXX"
echo "    PKG_CONFIG=$PKG_CONFIG"

# -ldflags "-s -w" and --tags=windowsqtstatic are required (TASK.md §Constraints).
# ./... builds every library package (compile + link verification); no main
# package exists yet, so no .exe is emitted.
go build \
    -gcflags=-trimpath="${GOPATH:-$HOME/go}" \
    -asmflags=-trimpath="${GOPATH:-$HOME/go}" \
    -ldflags "-s -w" \
    --tags=windowsqtstatic \
    ./...

echo "==> Library cross-compile OK (no main package, so no .exe emitted)."

# Once a `package main` entry point exists (e.g. gofluent/cmd/demo/main.go),
# replace the `go build ./...` command above with:
#
#   go build \
#       -gcflags=-trimpath="${GOPATH:-$HOME/go}" \
#       -asmflags=-trimpath="${GOPATH:-$HOME/go}" \
#       -ldflags "-s -w" \
#       --tags=windowsqtstatic \
#       -o gofluent.exe ./cmd/demo
