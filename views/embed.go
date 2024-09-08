package views

import "embed"

//go:embed home/*
//go:embed payments/*
//go:embed *
var FS embed.FS
