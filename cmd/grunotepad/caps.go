package main

// Explicit capability opt-in for the Notepad host.
import (
	_ "github.com/ledocorp/gru/ui/markdown"
	_ "github.com/ledocorp/gru/ui/spell"
	_ "github.com/ledocorp/gru/ui/syntax"
)
