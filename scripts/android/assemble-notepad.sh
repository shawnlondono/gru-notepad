#!/usr/bin/env bash
# Build libgrunotepad.so + assemble debug APK.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"

if [[ $# -eq 0 ]]; then
  "$ROOT/scripts/android/build-notepad.sh" arm64
  "$ROOT/scripts/android/build-notepad.sh" x86_64
else
  "$ROOT/scripts/android/build-notepad.sh" "$1"
fi

cd "$ROOT/packaging/android"
echo "sdk.dir=$ANDROID_HOME" > local.properties
./gradlew :notepad:assembleDebug --no-daemon

APK="$ROOT/packaging/android/notepad/build/outputs/apk/debug/notepad-debug.apk"
mkdir -p "$ROOT/dist/android"
cp -a "$APK" "$ROOT/dist/android/notepad-debug.apk"
ls -lh "$APK" "$ROOT/dist/android/notepad-debug.apk"
unzip -l "$APK" | grep 'lib/' || true
echo "Install: adb install -r dist/android/notepad-debug.apk"
echo "Launch:  adb shell am start -n org.ledocorp.gru.notepad/.MainActivity"
