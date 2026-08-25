# Gru Notepad v0.8.0

## Summary

**Gru Notepad** is a **product-shaped showcase** of the [Gru UI toolkit](https://github.com/ledocorp/gru): edge shell, text editor, multi-doc sessions, markdown preview, syntax highlight, and spell check in one thin desktop host.

**v0.8.0** refreshes the showcase under **GNU GPLv3**, pins the toolkit at **`github.com/ledocorp/gru@v0.8.0`**, and ships recent desktop polish (resize, preview, settings, and narrow-width chrome).

This is a **capability template** at 0.8.x — not “Notepad product v1.”

## What's new since v0.7.2

- **License:** **GNU GPLv3** (LedoCorp-published showcase)
- **Toolkit:** `require github.com/ledocorp/gru v0.8.0`
- **Resize / preview:** smoother window resize; markdown preview settles cleanly after drag
- **Settings:** more reliable desktop settings ribbon (including Back)
- **Narrow layout:** mobile-style bottom nav and Open Notes sheet stability
- **Host:** thin `cmd/grunotepad` sync; public-only comments and welcome copy

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

- **Toolkit:** [github.com/ledocorp/gru](https://github.com/ledocorp/gru) @ v0.8.0
- **Site:** [ledocorp.org/gru/notepad](https://ledocorp.org/gru/notepad)
- **Clips:** [X thread](https://x.com/ledocorp/status/2085492726074626160)

## License

GNU GPLv3 — see `LICENSE`. Official repo is **BDFL** (maintainer-led; no public PRs). Forks under GPLv3 are welcome.

---

*LedoCorp / Shawn Londono — Gru Notepad — GPLv3*
