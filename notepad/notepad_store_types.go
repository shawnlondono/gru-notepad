package notepad

import "github.com/ledocorp/gru/ui"

const (
	notepadSchemaVersion = 4
	notepadRecentMax     = 20
	notepadOpenDocsMax   = 20
)

// NotepadPrefs holds persisted scene preferences.
type NotepadPrefs struct {
	WordWrap          bool
	ShowStatus        bool
	SyntaxHighlight   bool
	SyntaxLangPick    string
	DefaultZoom       float32
	ShowSessionPanel  bool
	MobilePanelSheets bool // tablet/mobile: notes + preview as full-screen sheets
	MobileBottomNav   bool // tablet/mobile: icon bottom nav instead of top chrome
	SpellCheck        bool
	ChromeMode        string // "menubar" (default) or "ribbon"
	AppearanceMode    string // light, dark, system
	ChromaStyle       string // Chroma styles registry name
	FindUseModal      bool   // find opens modal; when false, ribbon search bar is used
	EditorBackground  string // "r,g,b,a" or empty for theme default
}

// OpenDocRow is one persisted open document (A4 session restore).
type OpenDocRow struct {
	DocID       string
	Path        string
	DisplayName string
	SortOrder   int
	ShowPreview bool
	IsActive    bool
	BufferText  string
	SavedText   string // last saved snapshot; buffer_text != saved_text => dirty
}

func defaultNotepadPrefs() NotepadPrefs {
	return NotepadPrefs{
		ShowStatus:        true,
		SyntaxHighlight:   true,
		SyntaxLangPick:    "auto",
		DefaultZoom:       1,
		ShowSessionPanel:  true,
		MobilePanelSheets: true,
		MobileBottomNav:   true,
		AppearanceMode:    ui.AppearanceSystem,
		ChromaStyle:       "",
		FindUseModal:      true,
	}
}
