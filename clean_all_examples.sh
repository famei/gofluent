#!/usr/bin/env bash
# clean_all_examples.sh — delete every .exe produced by build_all_examples.sh
# (each <dir>/<name>.exe placed next to its own main.go under examples/).
#
# Run inside WSL from the gofluent module directory:
#   wsl -e bash /mnt/c/.../gofluent/clean_all_examples.sh
set -uo pipefail
cd "$(dirname "$0")"

# Mirror build_all_examples.sh: every directory under examples/ with a main.go.
mapfile -t dirs < <(find examples -name main.go -printf '%h\n' | sort)

removed=0
for d in "${dirs[@]}"; do
  name="$(basename "$d")"
  exe="$d/$name.exe"
  if [ -f "$exe" ]; then
    rm -f "$exe"
    echo "removed $exe"
    removed=$((removed+1))
  fi
done

# Also drop the build log produced by build_all_examples.sh.
rm -f build_all_examples.log

echo "===== removed $removed exe file(s) ====="
