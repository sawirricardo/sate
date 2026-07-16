package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// A curated catalog: only translations whose license allows storing the
// full text locally (public domain, or explicitly free to distribute).
// Copyrighted-and-closed translations (e.g. LAI's Terjemahan Baru) can
// never be listed here — they would only be legal via per-day API fetches.
type translation struct {
	ID       string
	Name     string
	Language string
	Canon    string // "Catholic" includes the deuterocanon the lectionary cites
	License  string
	url      string
}

var catalog = []translation{
	{"dra", "Douay-Rheims 1899, American Edition", "English", "Catholic", "Public domain",
		"https://api.getbible.net/v2/douayrheims.json"},
	{"ayt", "Alkitab Yang Terbuka (YLSA/SABDA)", "Indonesian", "Protestant", "CopyLeft, non-commercial",
		"https://raw.githubusercontent.com/sabdacode/ayt/main/json/ayt.json"},
	{"web", "World English Bible", "English", "Protestant", "Public domain",
		"https://api.getbible.net/v2/web.json"},
	{"kjv", "King James Version 1769", "English", "Protestant", "Public domain",
		"https://api.getbible.net/v2/kjv.json"},
	{"asv", "American Standard Version 1901", "English", "Protestant", "Public domain",
		"https://api.getbible.net/v2/asv.json"},
	{"ylt", "Young's Literal Translation 1898", "English", "Protestant", "Public domain",
		"https://api.getbible.net/v2/ylt.json"},
}

// activeBibleID returns the translation chosen via `sate bible use`, or ""
// (callers then fall back to the first installed in catalog order).
func activeBibleID() string {
	raw, err := os.ReadFile(filepath.Join(dataDir(), "default"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
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
		active := activeBibleID()
		fmt.Println("Available translations (licenses allow local storage):")
		fmt.Printf("  %-5s %-11s %-11s %-25s %s\n", "ID", "LANG", "CANON", "LICENSE", "NAME")
		for _, t := range catalog {
			mark := " "
			if _, err := os.Stat(filepath.Join(dataDir(), t.ID+".json")); err == nil {
				mark = "*"
				if t.ID == active {
					mark = ">"
				}
			}
			fmt.Printf("%s %-5s %-11s %-11s %-25s %s\n", mark, t.ID, t.Language, t.Canon, t.License, t.Name)
		}
		fmt.Println("\n* = installed, > = active (switch with: sate bible use <id>).")
		fmt.Println("Protestant-canon bibles lack Wis/Sir/Bar/Macc, which the lectionary")
		fmt.Println("cites regularly — those readings fall back to citation-only.")
		return nil
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: sate bible use <id>")
		}
		id := args[1]
		if _, err := os.Stat(filepath.Join(dataDir(), id+".json")); err != nil {
			return fmt.Errorf("%s is not installed — run: sate bible add %s", id, id)
		}
		if err := os.WriteFile(filepath.Join(dataDir(), "default"), []byte(id), 0o644); err != nil {
			return err
		}
		fmt.Println("using", id)
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
	resp, err := http.Get(t.url)
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
	if t.ID == "ayt" {
		if raw, err = convertAYT(raw); err != nil {
			return err
		}
	}
	var probe struct {
		Books []struct {
			Name string `json:"name"`
		} `json:"books"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || len(probe.Books) == 0 {
		return fmt.Errorf("unexpected response format from %s", t.url)
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

// protestantOrder maps AYT's book numbers (standard 66-book order) to our
// canonical codes, whose English names read.go already resolves.
var protestantOrder = []string{
	"Gen", "Exod", "Lev", "Num", "Deut", "Josh", "Judg", "Ruth",
	"1Sam", "2Sam", "1Kgs", "2Kgs", "1Chr", "2Chr", "Ezra", "Neh", "Esth",
	"Job", "Ps", "Prov", "Eccl", "Cant",
	"Isa", "Jer", "Lam", "Ezek", "Dan",
	"Hos", "Joel", "Amos", "Obad", "Jonah", "Mic", "Nah", "Hab", "Zeph", "Hag", "Zech", "Mal",
	"Matt", "Mark", "Luke", "John", "Acts", "Rom", "1Cor", "2Cor",
	"Gal", "Eph", "Phil", "Col", "1Thess", "2Thess", "1Tim", "2Tim",
	"Titus", "Phlm", "Heb", "Jas", "1Pet", "2Pet", "1John", "2John", "3John", "Jude", "Rev",
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

// convertAYT reshapes SABDA's flat verse list into the nested books/
// chapters/verses format the reader expects, naming books in English so
// citation lookup works unchanged.
func convertAYT(raw []byte) ([]byte, error) {
	var flat []struct {
		Book    string `json:"book"`
		Chapter string `json:"chapter"`
		Verse   string `json:"verse"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal(raw, &flat); err != nil || len(flat) == 0 {
		return nil, fmt.Errorf("unexpected AYT format: %v", err)
	}
	type verse struct {
		Verse int    `json:"verse"`
		Text  string `json:"text"`
	}
	type chapter struct {
		Chapter int     `json:"chapter"`
		Verses  []verse `json:"verses"`
	}
	type book struct {
		Name     string     `json:"name"`
		Chapters []*chapter `json:"chapters"`
	}
	var books []*book
	byName := map[string]*book{}
	for _, v := range flat {
		bn, _ := strconv.Atoi(v.Book)
		if bn < 1 || bn > len(protestantOrder) {
			return nil, fmt.Errorf("book number %q out of range", v.Book)
		}
		name := bookNames[protestantOrder[bn-1]]
		b := byName[name]
		if b == nil {
			b = &book{Name: name}
			byName[name] = b
			books = append(books, b)
		}
		cn, _ := strconv.Atoi(v.Chapter)
		var ch *chapter
		if n := len(b.Chapters); n > 0 && b.Chapters[n-1].Chapter == cn {
			ch = b.Chapters[n-1]
		} else {
			ch = &chapter{Chapter: cn}
			b.Chapters = append(b.Chapters, ch)
		}
		vn, _ := strconv.Atoi(v.Verse)
		text := strings.TrimSpace(tagRe.ReplaceAllString(v.Text, " "))
		ch.Verses = append(ch.Verses, verse{vn, text})
	}
	return json.Marshal(map[string]any{"books": books})
}
