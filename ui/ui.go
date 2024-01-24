package ui

import "embed"

// Dist contains the static files for the UI.
//
//go:embed dist/*
var Dist embed.FS
