package notepad

// Shared spell-check extras and auto-correct table (dev stub + release Hunspell).
var notepadSpellExtraWords = []string{
	"notepad", "markdown", "gory", "gru", "utf", "api", "sql", "sqlite",
}

// notepadSpellAutoCorrect is the v1 replace-on-space table (lowercase keys).
var notepadSpellAutoCorrect = map[string]string{
	"teh":     "the",
	"recieve": "receive",
	"adn":     "and",
	"hte":     "the",
	"taht":    "that",
	"wiht":    "with",
	"thier":   "their",
}
