# Android packaging — Gru Notepad

Beta NativeActivity host for Gru Notepad. Toolkit Android APIs live in **GRU**
(`android-experience` until published). This repo only owns the product host.

| Module | Shared lib | Application id |
|--------|------------|----------------|
| `notepad/` | `libs/*/libgrunotepad.so` | `org.ledocorp.gru.notepad` |

**Persistence:** JSON `store.json` under app files (sqlite stays desktop-only for now).

## Local toolkit wiring

```bash
# In go.mod (local only — do not commit):
#   replace github.com/ledocorp/gru => ../GRU
# Checkout GRU branch android-experience as a sibling ../GRU
```

`scripts/android/env.sh` sources `../GRU/scripts/android/env.sh` for JDK/SDK/NDK.

## Build

```bash
source scripts/android/env.sh
./scripts/android/assemble-notepad.sh x86_64
adb install -r dist/android/notepad-debug.apk
adb shell am start -n org.ledocorp.gru.notepad/.MainActivity
./scripts/android/scrcpy-notepad.sh    # Cosmic-safe mirror
```

MainActivity: system-bar tint, IME EditText proxy, SAF open/save, `openGruUrl` (preview links).
