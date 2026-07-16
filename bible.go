package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// A curated catalog: public-domain translations only, so `add` may store the
// full text locally. Copyrighted translations (e.g. LAI's Terjemahan Baru)
// can never be listed here — they are only legal via per-day API fetches.
type translation struct {
	ID       string
	Name     string
	Language string
	Canon    string // "Catholic" includes the deuterocanon the lectionary cites
}

var catalog = []translation{
	{"dra", "Douay-Rheims 1899, American Edition", "English", "Catholic"},
	{"web", "World English Bible", "English", "Protestant"},
	{"kjv", "King James Version 1769", "English", "Protestant"},
	{"asv", "American Standard Version 1901", "English", "Protestant"},
	{"ylt", "Young's Literal Translation 1898", "English", "Protestant"},
}

func dataDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		return ".sate"
	}
	return filepath.Join(dir, ".local", "share", "sate")
}

func bibleCmd(args []string) error {
	sub := "ls"
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "ls":
		fmt.Println("Public-domain translations:")
		fmt.Printf("  %-5s %-9s %-11s %s\n", "ID", "LANG", "CANON", "NAME")
		for _, t := range catalog {
			mark := " "
			if _, err := os.Stat(filepath.Join(dataDir(), t.ID+".json")); err == nil {
				mark = "*"
			}
			note := ""
			if t.Canon == "Catholic" {
				note = "  (covers all lectionary books)"
			}
			fmt.Printf("%s %-5s %-9s %-11s %s%s\n", mark, t.ID, t.Language, t.Canon, t.Name, note)
		}
		fmt.Println("\n* = installed. Protestant-canon bibles lack Wis/Sir/Bar/Macc,")
		fmt.Println("which the lectionary cites regularly — prefer dra.")
		fmt.Println("No Indonesian translation is public domain; Terjemahan Baru")
		fmt.Println("would need a per-user api.bible key (not yet supported).")
		return nil
	case "add", "rm":
		return fmt.Errorf("bible %s: not implemented yet", sub)
	default:
		return fmt.Errorf("usage: sate bible [ls|add <id>|rm <id>]")
	}
}
