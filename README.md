# Gru Notepad

**Showcase version:** 0.7.2 (aligned with GRU toolkit)  
**License:** [Mozilla Public License 2.0](LICENSE)  
**Depends on:** [`github.com/ledocorp/gru`](https://github.com/ledocorp/gru) `@v0.7.2` *(pin after toolkit tag; local dry-run may still resolve via `replace`)*

Gru Notepad is a **product-shaped example** of the Gru UI toolkit: edge shell, text editor, multi-doc sessions, markdown preview, syntax highlight, and spell check in one thin desktop host.

It is **not** “Notepad product v1.” Treat this as a **v0.7.x capability showcase / template**.

| | |
|--|--|
| **Publisher** | [LedoCorp](https://ledocorp.org) |
| **Site** | [ledocorp.org/gru/notepad](https://ledocorp.org/gru/notepad) *(planned)* |
| **Toolkit** | [`ledocorp/gru`](https://github.com/ledocorp/gru) (separate MPL repo) |

---

## Layout (this repo)

```text
cmd/grunotepad/     thin window host + capability blank-imports
notepad/            scene package (editor, preview, settings, store)
appicon/            public-boundary window/title icons (stubs until PNGs embed)
media/preview/      short showcase videos (+ local HTML preview)
```

This is a **standalone GitHub repository**, not a folder inside `ledocorp/gru`. The app `require`s the toolkit module.

---

## Build / run

```powershell
# From this repo (after go mod tidy):
go run ./cmd/grunotepad

# Or build:
go build -o GruNotepad.exe ./cmd/grunotepad
```

Spell check expects Hunspell dictionaries beside the binary in release zips (`dict/en_US/`).

---

## Showcase videos

Clips are hosted on X (not required in this git tree):

| Clip | Post |
|------|------|
| Thread / desktop shell | https://x.com/ledocorp/status/2085492726074626160 |
| Basics | https://x.com/ledocorp/status/2085492728503189630 |
| Flexible layout | https://x.com/ledocorp/status/2085492730872893843 |
| Syntax | https://x.com/ledocorp/status/2085492733272023217 |

Site draft (LedoCorp styling): [`../ledocorp-site/gru/notepad/`](../ledocorp-site/gru/notepad/)

```powershell
# from staging/ledocorp-site
python -m http.server 8765
# open http://127.0.0.1:8765/gru/notepad/
```

---

## Status

| Item | State |
|------|--------|
| Toolkit pin | `github.com/ledocorp/gru@v0.7.2` (published) |
| Build against published module | Verified on Windows |
| Public GitHub | Ready (`shawnlondono/gru-notepad`) |
| ledocorp.org `/gru/notepad` | Draft in `staging/ledocorp-site/`; deploy separately |

Window/title icons in this repo are stubs (`appicon/`); polish later if needed.
