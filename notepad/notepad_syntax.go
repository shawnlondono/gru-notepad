package notepad

import (
	"path/filepath"
	"strings"

	"github.com/ledocorp/gru/ui"
)

// notepadSyntaxLanguage maps a file path to a Chroma lexer name (empty = plain text).
func notepadSyntaxLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".md", ".markdown":
		return "markdown"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".sql":
		return "sql"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx":
		return "cpp"
	case ".java":
		return "java"
	case ".lua":
		return "lua"
	default:
		return ""
	}
}

// notepadManualSyntaxLanguages lists explicit Syntax menu picks (keep in sync with
// notepadSyntaxLanguage extensions so Auto and manual picks match).
func notepadManualSyntaxLanguages() []struct{ Label, Pick string } {
	return []struct{ Label, Pick string }{
		{"Markdown", "markdown"},
		{"Go", "go"},
		{"Python", "python"},
		{"Rust", "rust"},
		{"JavaScript", "javascript"},
		{"TypeScript", "typescript"},
		{"JSON", "json"},
		{"YAML", "yaml"},
		{"SQL", "sql"},
		{"HTML", "html"},
		{"CSS", "css"},
		{"Bash", "bash"},
		{"C", "c"},
		{"C++", "cpp"},
		{"Java", "java"},
		{"Lua", "lua"},
	}
}

func (s *notepadScene) syntaxLanguageMenuItems(includeAuto bool) []ui.ContextMenuItem {
	var items []ui.ContextMenuItem
	if includeAuto {
		items = append(items, ui.ContextMenuItem{
			Label:   "Auto (file name)",
			Checked: s.syntaxHighlight && (s.syntaxLangPick == "" || s.syntaxLangPick == "auto"),
			Action:  func() { s.setSyntaxLangPick("auto") },
		})
	}
	for _, e := range notepadManualSyntaxLanguages() {
		pick := e.Pick
		items = append(items, ui.ContextMenuItem{
			Label:   e.Label,
			Checked: s.syntaxHighlight && s.syntaxLangPick == pick,
			Action:  func() { s.setSyntaxLangPick(pick) },
		})
	}
	items = append(items, ui.ContextMenuItem{
		Label:   "Plain Text",
		Checked: s.syntaxHighlight && s.syntaxLangPick == "plain",
		Action:  func() { s.setSyntaxLangPick("plain") },
	})
	return items
}
