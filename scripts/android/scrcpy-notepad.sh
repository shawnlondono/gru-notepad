#!/usr/bin/env bash
# Mirror headless AVD showing Gru Notepad (Cosmic-safe).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
GRU_ROOT="$(cd "$ROOT/../GRU" && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"
export PATH="$ANDROID_HOME/platform-tools:$PATH"

SCRCPY="$GRU_ROOT/.android/scrcpy/scrcpy"
if [[ ! -x "$SCRCPY" ]]; then
  echo "scrcpy missing under $GRU_ROOT/.android/scrcpy/" >&2
  exit 1
fi

pkill -f '/.android/scrcpy/scrcpy' 2>/dev/null || true
sleep 0.5

adb wait-for-device
adb shell am start -n org.ledocorp.gru.notepad/.MainActivity >/dev/null 2>&1 || true

export WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-1}"
unset DISPLAY || true
export SDL_VIDEODRIVER=wayland

exec "$SCRCPY" \
  --window-title "Android Emulator (notepad)" \
  --video-bit-rate 20M \
  --max-fps 60 \
  --no-audio \
  "$@"
