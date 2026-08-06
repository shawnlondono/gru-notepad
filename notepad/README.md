# notepad

Scene package for **Gru Notepad**: editor, Open Notes session, markdown preview, settings, SQLite store, and chrome.

## Run

From the repo root:

```powershell
go run ./cmd/grunotepad
```

## Layout

| Path | Role |
|------|------|
| `register.go` | `SceneTitle`, `NewScene()` |
| `notepad_*.go` | Scene, chrome, session, preview, store, caps wiring |
| `settings_rows.go` | Flat settings row helpers used by desktop settings |

Shared shell helpers come from the GRU toolkit (`examples`: ListPane, edge root, filedialogs).
