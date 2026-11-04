#!/usr/bin/env bash
# Build cmd/grunotepad as a NativeActivity shared library for Android.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/android/env.sh"

ARCH="${1:-arm64}" # arm64 | x86_64

ANDROID_API="${ANDROID_API:-21}"
PREBUILT="${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64"
SYSROOT="${PREBUILT}/sysroot"
export PATH="${PREBUILT}/bin:${PATH}"

case "$ARCH" in
  arm64|aarch64)
    GOARCH=arm64
    TRIPLE=aarch64-linux-android
    ABI=arm64-v8a
    ;;
  x86_64|x64|amd64)
    GOARCH=amd64
    TRIPLE=x86_64-linux-android
    ABI=x86_64
    ;;
  *)
    echo "usage: $0 [arm64|x86_64]" >&2
    exit 1
    ;;
esac

OUTDIR="$ROOT/packaging/android/notepad/libs/${ABI}"
mkdir -p "$OUTDIR"

# Stage toolkit fonts/icons into APK assets (via assets/ symlink → ../GRU/assets).
ASSETS="$ROOT/packaging/android/notepad/assets"
mkdir -p "$ASSETS"
for d in fonts icons; do
  if [[ -d "$ROOT/assets/$d" ]]; then
    mkdir -p "$ASSETS/$d"
    cp -a "$ROOT/assets/$d/." "$ASSETS/$d/"
  fi
done

GLUE_INC="${ANDROID_NDK_HOME}/sources/android/native_app_glue"

echo "ANDROID_NDK_HOME=$ANDROID_NDK_HOME"
echo "Building libgrunotepad.so ($ABI, API $ANDROID_API)..."

CC="${TRIPLE}${ANDROID_API}-clang" \
CGO_CFLAGS="-I${SYSROOT}/usr/include -I${SYSROOT}/usr/include/${TRIPLE} -I${GLUE_INC} --sysroot=${SYSROOT}" \
CGO_LDFLAGS="-L${SYSROOT}/usr/lib/${TRIPLE}/${ANDROID_API} --sysroot=${SYSROOT}" \
CGO_ENABLED=1 GOOS=android GOARCH="$GOARCH" \
go build -buildmode=c-shared \
  -ldflags="-s -w -extldflags=-Wl,-soname,libgrunotepad.so" \
  -o="${OUTDIR}/libgrunotepad.so" \
  "$ROOT/./cmd/grunotepad"

rm -f "${OUTDIR}/libgrunotepad.h"
ls -lh "${OUTDIR}/libgrunotepad.so"
file "${OUTDIR}/libgrunotepad.so"
echo "OK — next: ./scripts/android/assemble-notepad.sh"
