package notepad

// Welcome pack — seeded open notes for a clean Gru Notepad install.
// Meta key welcome_pack gates one-shot wipe+seed (see NotepadStore.EnsureWelcomePack).

const notepadWelcomePackVersion = "6"

type welcomeNote struct {
	id          string
	displayName string
	showPreview bool
	text        string
}

func notepadWelcomeNotes() []welcomeNote {
	// Preview stays off until the user toggles it; switching notes never forces it on.
	return []welcomeNote{
		{
			id:          "welcome-gru",
			displayName: "Welcome · Gru",
			showPreview: false,
			text:        welcomeGruMD(),
		},
		{
			id:          "welcome-notepad",
			displayName: "Gru Notepad",
			showPreview: false,
			text:        welcomeNotepadMD(),
		},
		{
			id:          "welcome-fx",
			displayName: "fx · LedoCorp",
			showPreview: false,
			text:        welcomeFxMD(),
		},
		{
			id:          "welcome-syntax",
			displayName: "Syntax sample",
			showPreview: false,
			text:        welcomeSyntaxGo(),
		},
		{
			id:          "welcome-markdown",
			displayName: "Markdown gallery",
			showPreview: false,
			text:        welcomeMarkdownGalleryMD(),
		},
	}
}

func welcomeGruMD() string {
	return `# Welcome to Gru

**Gru** is LedoCorp's retained-mode UI toolkit for native desktop apps. You write Go scenes, compose widgets by hand, and ship a thin host: no React-in-Go layer, no hidden virtual DOM.

> This Notepad build is a **capability showcase** (v0.7.x era): editor, preview, syntax, and spell in one product-shaped host. It is not a finished commercial notes suite.

## What “retained-mode” means here

Widgets keep identity across frames. Layout and draw mark dirty regions instead of rebuilding the tree every tick. That is how Gru stays responsive with split panes, ribbons, and markdown preview open at once.

## Why Gru

1. **Pay for what you import**: lean hosts stay lean; caps (syntax, spell, markdown) opt in at link time
2. **Composition over ceremony**: build with flex plus ` + "`Panel`" + ` / ` + "`Card`" + ` / ` + "`ListTile`" + `, not a second framework
3. **Same engine, many hosts**: Hello, Calculator, and **Gru Notepad** share one ` + "`ui`" + ` spine
4. **Desktop-first chrome**: MenuBar, ribbon, status bar, edge-to-edge root

## Quick map

| Piece | Role |
|-------|------|
| ` + "`ui`" + ` | Widgets, flex layout, chrome, theme, type scale |
| ` + "`ui/syntax`" + ` | Chroma highlighting (blank-import capability) |
| ` + "`ui/spell`" + ` | Hunspell + dictionaries |
| ` + "`ui/preview`" + ` | GFM markdown (` + "`MarkdownView`" + `) |
| Thin hosts | ` + "`cmd/hello`" + `, ` + "`cmd/calc`" + `, ` + "`cmd/grunotepad`" + ` |

## How apps are shaped

A typical host:

1. Opens a window and creates a ` + "`ui.Document`" + `
2. Mounts a shell (` + "`examples.MountEdgeToEdgeRoot`" + ` for desktop Notepad)
3. Builds scenes: panels, lists, editors, modals
4. Optionally enables caps so syntax / spell / preview link in

## Learn more

- Toolkit: [ledocorp/GRU](https://github.com/ledocorp/gru) on GitHub
- LedoCorp: [ledocorp.org](https://ledocorp.org)
- Sibling language: [ledocorp.org/fx](https://ledocorp.org/fx)

---

Open **Gru Notepad**, **fx**, **Syntax sample**, and **Markdown gallery** in the sidebar for more.
`
}

func welcomeNotepadMD() string {
	return `# Gru Notepad

A small notes app built on Gru, meant to show how several toolkit capabilities work together in one product-shaped host. It is the first showcase app on the public toolkit path (v0.7.x capability template), not a claim that “Notepad product v1” is finished.

## What you have open

The **Open Notes** sidebar is a multi-document session: untitled buffers, welcome notes, and files you open stay listed until you close them. Switching notes keeps your preview toggle (preview does not auto-open when you click a note).

SQLite under the hood remembers open docs, recent files, and prefs (wrap, spell, theme, chrome mode, and more).

## Try these

- [ ] Toggle **View → Markdown Preview** on a markdown note
- [ ] Open **Markdown gallery** for tables, tasks, quotes, and fences
- [ ] Open **Syntax sample** → **Syntax → Highlight** (Auto / Go / plain)
- [ ] **File → Settings** — theme, syntax theme (Chroma), wrap, spell, chrome
- [ ] Resize narrow → ribbon + optional mobile bottom nav; widen → desktop chrome
- [ ] **Ctrl+F** Find in the ribbon (or modal, per Settings)

## Shortcuts

| Action | Keys |
|--------|------|
| New | Ctrl+N |
| Open | Ctrl+O |
| Save | Ctrl+S |
| Save As | Ctrl+Shift+S |
| Find | Ctrl+F |
| Replace | Ctrl+H |
| Zoom in / out / reset | Ctrl++ / Ctrl+- / Ctrl+0 |
| Find next / previous | F3 / Shift+F3 |

## What this shows off

| Capability | Where |
|------------|--------|
| Edge shell + menubar / ribbon / status | Chrome |
| Sessions / Open Notes list pane | Sidebar |
| Hunspell spell check + autocorrect table | Edit + Settings |
| Chroma syntax (language + theme) | Syntax menu + Settings |
| GFM markdown preview | View → Markdown Preview |
| Find / replace | Search ribbon + modals |
| Appearance light / dark / system | Settings → Theme |

## Caps and the thin host

` + "`cmd/grunotepad`" + ` links optional capabilities (markdown, syntax, spell) at build time. The scene package lives beside the host in this repository and depends on the public GRU toolkit module.

### Responsive

On a narrow window, Notepad can switch to ribbon + mobile bottom nav and sheet-style Open Notes. Widen the window to return to MenuBar + desktop split panes.

### Honest scope

Ship-quality editor + preview for showcase and dogfood. Not a sync cloud, not a plugin marketplace, not Joplin/Obsidian feature parity. RAM can sit in the hundreds of MB with SSAA, atlases, and preview — that is an accepted quality trade for this host.

---

*Built with GRU · explore the caps, then scaffold your own app with the toolkit docs.*
`
}

func welcomeFxMD() string {
	return `# fx

**fx** is LedoCorp's systems language — a sibling project to Gru, not bundled inside this Notepad. It refuses hidden control: no garbage collector quietly rewriting your heap, no implicit allocator, no framework magic between what you wrote and what runs.

**Site:** [ledocorp.org/fx](https://ledocorp.org/fx)

**Repo:** [github.com/ledocorp/fxlang](https://github.com/ledocorp/fxlang)

**License:** Apache-2.0 · current line **0.8.x**

## What it is

fx sits on top of the C universe without pretending C never existed. You get a sharper language for new code, and a first-class path back to readable C when you need to ship, embed, or wrap what you already have.

Every important decision — who allocates, who mutates, how long memory lives — is meant to stay visible in the source you open tomorrow, or that an agent generates tonight.

## Core ideas

- **IR → native by default** — everyday ` + "`fx run`" + ` / ` + "`fx build`" + ` lower through IR
- **Emit-C stays first-class** — ` + "`--emit-c`" + `, annotate comments, ` + "`.fxmap`" + `, ` + "`fx locate`" + `
- **Effects** — signatures declare ` + "`alloc`" + ` / ` + "`mut`" + ` so cost stays local
- **Named regions** — arena lifetimes you can see; batch free when the region ends (not a GC)
- **Wrap / host** — C owns ` + "`main`" + `, fx owns the logic; link mature C libraries instead of rewriting them
- **Small zspec substrate** — explicit allocators and errors in the linked C world
- **Humans and agents** — low-ceremony, mechanical style that stays auditable when generated

## Who it is for

- Systems and embedded work that needs structure without surrendering the heap to a GC
- Teams extending codebases that already speak C
- People using AI assistants who still want inspectable, explicit programs

It is **not** a batteries-included app platform or a scripting GC language. The 0.8 line leads with IR → native, keeps readable emit-C, and ships dogfood tools — still with no silent “just mutate the slice” dialect.

## Try it

Requirements: the language package from GitHub, a C toolchain (` + "`gcc`" + ` default), Windows or Linux x86_64.

` + "```bash" + `
# put bin/ on PATH, then:
fx doctor                  # check gcc + zspec paths
fx version
fx new hello
cd hello
fx run main.fx             # IR → native
fx run main.fx --emit-c    # optional: emit-C → native
` + "```" + `

More scaffolds and host recipes:

` + "```bash" + `
fx new tiny --scaffold minimal
fx new firmware --scaffold embedded
fx emit-c main.fx -o out_c
fx run examples/showcase_wrap/compute.fx --host examples/showcase_wrap/host.c
` + "```" + `

## What is in the package

| Path | Role |
|------|------|
| ` + "`bin/fx`" + ` | Compiler / runner |
| ` + "`scaffolds/`" + `, ` + "`std/`" + `, ` + "`lib/`" + ` | Starters and libraries |
| ` + "`docs/`" + `, ` + "`examples/`" + ` | Guides and dogfood |
| ` + "`zspec/`" + ` + ` + "`build/`" + ` | Headers / libzspec for emitted C |
| ` + "`compiler-source/`" + ` | Read-only fx modules for inspection |

## Tooling

The same binary includes ` + "`fx doctor`" + ` (toolchain check), ` + "`fx lsp`" + ` (editor), and ` + "`fx mcp`" + ` (lean agent tools: check / locate / run). Put ` + "`fx`" + ` on your ` + "`PATH`" + ` and wire Cursor or friends as you like.

## Recent 0.8.x themes

- IR by default for everyday run/build
- Dual-emit tracking (annotate comments, ` + "`.fxmap`" + `, ` + "`#line`" + `, ` + "`fx locate`" + `)
- Dogfood apps (CHIP‑8, JSON wrap + pure, line-LCS textdiff)
- Clearer diagnostics and thin C CLI-host recipes
- Composition under regions (typed pool ids, ` + "`vec_set`" + `, teaching patterns)

Full notes ship in the language package as ` + "`docs/releases/`" + ` and on GitHub Releases when tagged.

> Gru and fx share LedoCorp DNA (explicit control, inspectable systems) but different jobs: Gru is the UI toolkit; fx is the systems language. This note is a pointer for explorers.
`
}

func welcomeSyntaxGo() string {
	return `// Gru Notepad — syntax highlight sample (Go)
// Turn on: Syntax → Highlight  (language: Auto or Go)

package main

import (
	"github.com/ledocorp/gru/examples"
	"fmt"
	"strings"
)

// fib returns the n-th Fibonacci number (n >= 0).
func fib(n int) int {
	if n < 2 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {
	title := strings.ToUpper("gru notepad")
	fmt.Println(title)
	for i := 0; i < 10; i++ {
		fmt.Printf("fib(%d) = %d\n", i, fib(i))
	}
}
`
}

func welcomeMarkdownGalleryMD() string {
	return `# Markdown gallery

This note is a **GFM smoke test** for Gru's ` + "`preview.MarkdownView`" + `. Toggle preview and resize the window — headings, tables, and code should reflow cleanly.

## Emphasis

You can mix *italic*, **bold**, ***both***, and ` + "`inline code`" + `.

## Lists

Unordered:

- Edge-to-edge shell
- Split panes (editor | preview)
- Optional caps at link time

Ordered:

1. Clone / build against tagged GRU
2. Open a note
3. Turn on preview

Nested:

- File
  - New / Open / Save
  - Settings
- View
  - Markdown Preview
  - Zoom

## Task list

- [x] Ship thin ` + "`cmd/grunotepad`" + ` host
- [x] Welcome pack with real content
- [ ] Public ` + "`gru-notepad`" + ` staging dry-run
- [ ] Pin showcase to GRU v0.7.x

## Table

| Widget | Package | Notes |
|--------|---------|-------|
| TextEditor | ` + "`ui`" + ` | Zoom, wrap, spell |
| MarkdownView | ` + "`ui/preview`" + ` | GFM + footnotes |
| Syntax | ` + "`ui/syntax`" + ` | Chroma blank-import |
| Spell | ` + "`ui/spell`" + ` | Hunspell + dict/ |

## Quote

> Composition patterns beat one-off layout stacks.  
> — Gru cookbook spirit

## Code

Go:

` + "```go" + `
func Hello(name string) string {
	return "Hello, " + name
}
` + "```" + `

JSON:

` + "```json" + `
{
  "product": "Gru Notepad",
  "toolkit": "gru",
  "version": "0.7.x"
}
` + "```" + `

## Rule

---

## Links

- [ledocorp.org](https://ledocorp.org)
- [ledocorp.org/fx](https://ledocorp.org/fx)
- [GRU on GitHub](https://github.com/ledocorp/gru)

Footnote demo:[^1]

[^1]: Preview supports GFM footnotes when the pack is built with the footnote extension.
`
}
