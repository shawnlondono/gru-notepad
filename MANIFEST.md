# Gru Notepad — public manifest (IN / OUT)

**Target repo:** `github.com/shawnlondono/gru-notepad`  
**License:** MPL-2.0  
**Toolkit pin:** `github.com/ledocorp/gru@v0.7.1` (bump only if HEAD APIs require it)  
**Parent plan:** [`NOTEPAD_PUBLIC_PLAN.md`](../NOTEPAD_PUBLIC_PLAN.md)

## Framing

Public Notepad is a **GRU capability showcase / example template**, versioned in the **0.7.x** band to match the toolkit — not a claim of Notepad product v1.

## IN (export allowlist — draft)

| Path (monorepo → public) | Notes |
|--------------------------|--------|
| `cmd/grunotepad/` | Thin host only (no `notepad` build tag in public repo) |
| `notepad/` | Scene package (`apps/notepad` in monorepo) |
| `appicon/` | Public-boundary icon helpers (not `gru/internal/appicon`) |
| `media/preview/` | Showcase videos + local HTML preview page |
| Minimal glue | Via toolkit `examples` (list pane / settings / filedialogs) |
| Fonts + `dict/` layout docs | Ship or document sidecar |
| `LICENSE` `NOTICE` `README` | MPL + third-party; README must credit GRU |

## OUT (never publish)

| Path / topic | Why |
|--------------|-----|
| `foundry/`, `apps/prism`, `devtools/studio` | Private |
| Root Studio `main.go`, grudemo catalog | Not the product |
| `docs/NOTEPAD_*`, BUILD_LOG, COMPOSITION_PATTERNS §11 | Internal engineering |
| GoRy / Studio planning under `staging/` except this tree’s public drafts | Keep pipelines separate |
| Full `examples/` demo catalog | Showcase is Notepad, not Studio |

## Guard checks (before push)

```text
rg -n "foundry|apps/prism|devtools/studio|COMPOSITION_PATTERNS|NOTEPAD_ROADMAP|NOTEPAD_BUILD_LOG" .
# expect: no hits in the export tree

go list -deps ./cmd/grunotepad | rg "foundry|prism|devtools/studio"
# expect: empty
```

## Version

| Artifact | Version |
|----------|---------|
| First public Notepad tag (proposed) | `v0.7.1` (showcase aligned with GRU) or `v0.1.0` if we want independent semver — **prefer documenting “showcase 0.7.1 / built on GRU v0.7.1”** in README either way |
| GRU require | `@v0.7.1` |

## Notice policy

- Notepad NOTICE: short product header + “see GRU NOTICE for toolkit deps” + any Notepad-only deps (sqlite, etc. if not covered).
- Do not paste private monorepo paths into public README.
