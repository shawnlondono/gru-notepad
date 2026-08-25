# Gru Notepad

**Showcase version:** 0.8.0 (aligned with GRU toolkit)  
**License:** [GNU GPLv3](LICENSE)  
**Depends on:** [`github.com/ledocorp/gru`](https://github.com/ledocorp/gru) `@v0.8.0`

Gru Notepad is a **product-shaped example** of the Gru UI toolkit: edge shell, text editor, multi-doc sessions, markdown preview, syntax highlight, and spell check in one thin desktop host.

It is **not** “Notepad product v1.” Treat this as a **v0.8.x capability showcase / template**.

| | |
|--|--|
| **Publisher** | [LedoCorp](https://ledocorp.org) |
| **Site** | [ledocorp.org/gru/notepad](https://ledocorp.org/gru/notepad) |
| **Toolkit** | [`ledocorp/GRU`](https://github.com/ledocorp/GRU) (separate GPLv3 repo) |

---

## Layout

```text
cmd/grunotepad/     thin window host + capability blank-imports
notepad/            scene package (editor, preview, settings, store)
appicon/            public-boundary window/title icons (stubs until PNGs embed)
media/preview/      clip index (videos hosted on X)
```

This is a **standalone GitHub repository**, not a folder inside `ledocorp/gru`. The app `require`s the toolkit module.

---

## Build / run

```powershell
go run ./cmd/grunotepad

# Or build:
go build -o GruNotepad.exe ./cmd/grunotepad
```

Spell check expects Hunspell dictionaries beside the binary in release zips (`dict/en_US/`).

---

## License

Gru Notepad is **GNU GPLv3**. The Gru toolkit it depends on is also GPLv3. Official trees are **BDFL** (no public PRs). LedoCorp may offer a commercial license later for proprietary use; that path is separate from this open-source tree.

---

## Showcase videos

Clips are hosted on X:

| Clip | Post |
|------|------|
| Thread / desktop shell | https://x.com/ledocorp/status/2085492726074626160 |
| Basics | https://x.com/ledocorp/status/2085492728503189630 |
| Flexible layout | https://x.com/ledocorp/status/2085492730872893843 |
| Syntax | https://x.com/ledocorp/status/2085492733272023217 |
