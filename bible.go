package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	gbID     string // id on api.getbible.net, the download source
}

var catalog = []translation{
	{"dra", "Douay-Rheims 1899, American Edition", "English", "Catholic", "douayrheims"},
	{"web", "World English Bible", "English", "Protestant", "web"},
	{"kjv", "King James Version 1769", "English", "Protestant", "kjv"},
	{"asv", "American Standard Version 1901", "English", "Protestant", "asv"},
	{"ylt", "Young's Literal Translation 1898", "English", "Protestant", "ylt"},
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
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: sate bible add <id>")
		}
		return addBible(args[1])
	case "rm":
		if len(args) < 2 {
			return fmt.Errorf("usage: sate bible rm <id>")
		}
		path := filepath.Join(dataDir(), args[1]+".json")
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("bible rm: %w", err)
		}
		fmt.Println("removed", args[1])
		return nil
	default:
		return fmt.Errorf("usage: sate bible [ls|add <id>|rm <id>]")
	}
}

func addBible(id string) error {
	var t *translation
	for i := range catalog {
		if catalog[i].ID == id {
			t = &catalog[i]
		}
	}
	if t == nil {
		return fmt.Errorf("unknown translation %q — see sate bible ls", id)
	}
	fmt.Printf("downloading %s (%s)...\n", t.ID, t.Name)
	resp, err := http.Get("https://api.getbible.net/v2/" + t.gbID + ".json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var probe struct {
		Books []struct {
			Name string `json:"name"`
		} `json:"books"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || len(probe.Books) == 0 {
		return fmt.Errorf("unexpected response format from getbible.net")
	}
	if err := os.MkdirAll(dataDir(), 0o755); err != nil {
		return err
	}
	path := filepath.Join(dataDir(), t.ID+".json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return err
	}
	fmt.Printf("installed %s: %d books, %.1f MB -> %s\n",
		t.ID, len(probe.Books), float64(len(raw))/1e6, path)
	return nil
}
