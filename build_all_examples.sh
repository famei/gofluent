#!/usr/bin/env bash
# build_all_examples.sh — compile EVERY example under examples/ and place each
# .exe next to its own main.go (same directory).
#
# Run inside WSL from the gofluent module directory:
#   wsl -e bash /mnt/c/.../gofluent/build_all_examples.sh
#
# Failures (e.g. examples that need unavailable Qt modules like QtMultimedia /
# QtWebEngine) are logged to build_all_examples.log and do not stop the batch.
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

# Collect every directory under examples/ that contains a main.go.
mapfile -t dirs < <(find examples -name main.go -printf '%h\n' | sort)

: > build_all_examples.log
ok=0
fail=0

for d in "${dirs[@]}"; do
  name="$(basename "$d")"
  pkg="./$d"
  out="$d/$name.exe"
  echo "==> build $pkg -> $out"
  if go build -ldflags "-s -w" --tags=windowsqtstatic -o "$out" "$pkg" >> build_all_examples.log 2>&1; then
    echo "    OK"
    ok=$((ok+1))
  else
    echo "    FAIL ($pkg)"
    fail=$((fail+1))
  fi
done

echo "===== done: ok=$ok fail=$fail ====="
echo "===== last 40 lines of build_all_examples.log ====="
tail -n 40 build_all_examples.log
