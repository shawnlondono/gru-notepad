package main

// Explicit capability opt-in for the Notepad host (MASTER_SEPARATION_PLAN Phase 2).
// Do not rely on examples/optional_capabilities.go — that aggregator is Studio/showcase only.
import (
	_ "github.com/ledocorp/gru/ui/markdown"
	_ "github.com/ledocorp/gru/ui/spell"
	_ "github.com/ledocorp/gru/ui/syntax"
)
