# Gru Notepad v0.9.0

## Summary

**Gru Notepad** is a **product-shaped showcase** of the [Gru UI toolkit](https://github.com/ledocorp/gru): edge shell, text editor, multi-doc sessions, markdown preview, syntax highlight, and spell check in one thin desktop host.

**v0.9.0** pins the toolkit at **`github.com/ledocorp/gru@v0.9.1`** (S1 ship default: 1× present + sticky text) so the public showcase matches the living toolkit tip.

This is a **capability template** at 0.9.x — not “Notepad product v1.”

## What's new since v0.8.0

- **Toolkit:** `require github.com/ledocorp/gru v0.9.1`
- **Present path:** inherits S1 defaults from the toolkit (1× framebuffer present; sticky ClearType/LCD Latin; optional `GRU_SSAA` / `GRU_RENDER_QUALITY` for testing)
- **Showcase:** same edge-shell product surface; version aligned with toolkit tip sync

## Install / run

```bash
git clone https://github.com/shawnlondono/gru-notepad
cd gru-notepad
go run ./cmd/grunotepad
```

Or build:

```bash
go build -o GruNotepad.exe ./cmd/grunotepad
```

Spell check expects Hunspell dictionaries beside the binary in release layouts (`dict/en_US/`).

## Links

- **Toolkit:** [github.com/ledocorp/gru](https://github.com/ledocorp/gru) @ v0.9.1
- **Site:** [ledocorp.org/gru/notepad](https://ledocorp.org/gru/notepad)
- **Clips:** [X thread](https://x.com/ledocorp/status/2085492726074626160)

## License

GNU GPLv3 — see `LICENSE`. Official repo is **BDFL** (maintainer-led; no public PRs). Forks under GPLv3 are welcome.

---

*LedoCorp / Shawn Londono — Gru Notepad — GPLv3*
