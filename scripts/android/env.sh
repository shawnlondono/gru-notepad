#!/usr/bin/env bash
# Reuse the GRU toolkit Android toolchain env when co-developing locally.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GRU_ROOT="$(cd "$HERE/../../../GRU" && pwd)"
if [[ -f "$GRU_ROOT/scripts/android/env.sh" ]]; then
  # shellcheck disable=SC1091
  source "$GRU_ROOT/scripts/android/env.sh"
else
  echo "Expected sibling GRU checkout at $GRU_ROOT with scripts/android/env.sh" >&2
  exit 1
fi
