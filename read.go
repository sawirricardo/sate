package main

// Verse lookup in an installed bible, used to expand the banner from
// citations to full reading texts when one is available.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sawirricardo/sate/scripture"
)

type bibleFile struct {
	id    string
	Books []struct {
		Name     string `json:"name"`
		Chapters []struct {
			Chapter int `json:"chapter"`
			Verses  []struct {
				Verse int    `json:"verse"`
				Text  string `json:"text"`
			} `json:"verses"`
		} `json:"chapters"`
	} `json:"books"`
}

// loadInstalledBible returns the first installed translation in catalog
// order (dra first — full Catholic canon), or nil when none is installed.
func loadInstalledBible() *bibleFile {
	for _, t := range catalog {
		raw, err := os.ReadFile(filepath.Join(dataDir(), t.ID+".json"))
		if err != nil {
			continue
		}
		var b bibleFile
		if json.Unmarshal(raw, &b) != nil || len(b.Books) == 0 {
			continue
		}
		b.id = t.ID
		return &b
	}
	return nil
}

// passage returns the joined verse text for refs, or ok=false when the
// translation cannot serve it (missing book, unmappable range).
func (b *bibleFile) passage(refs []scripture.Ref) (string, bool) {
	var parts []string
	for _, ref := range refs {
		name := bookNames[ref.Book]
		chapter := ref.Chapter
		if b.id == "dra" && ref.Book == "Ps" {
			chapter = vulgatePsalm(chapter)
		}
		var texts []string
		for i := range b.Books {
			if b.Books[i].Name != name {
				continue
			}
			for _, ch := range b.Books[i].Chapters {
				if ch.Chapter != chapter {
					continue
				}
				for _, v := range ch.Verses {
					if v.Verse >= ref.From && v.Verse <= ref.To {
						texts = append(texts, strings.TrimSpace(v.Text))
					}
				}
			}
			break
		}
		if len(texts) == 0 {
			return "", false // book not in this canon, or bad mapping
		}
		parts = append(parts, texts...)
	}
	return strings.Join(parts, " "), len(parts) > 0
}

// vulgatePsalm maps the Hebrew psalm numbering used by the lectionary to
// the Vulgate numbering the Douay-Rheims keeps.
// ponytail: split psalms map to their first half — H116:10-19 (V115) and
// H147:12-20 (V147) come back empty; add dual-chapter handling if it bites.
func vulgatePsalm(n int) int {
	switch {
	case n <= 8 || n >= 148:
		return n
	case n == 9 || n == 10:
		return 9
	case n <= 113:
		return n - 1
	case n == 114 || n == 115:
		return 113
	case n == 116:
		return 114
	case n <= 146:
		return n - 1
	default:
		return 146
	}
}

// printPassage prints one titled reading section with wrapped text.
func printPassage(b *bibleFile, label, citation, cycle string) {
	citation = scripture.PickCycle(citation, cycle)
	fmt.Printf("\n%s — %s\n", label, citation)
	refs, err := scripture.Parse(citation)
	if err != nil {
		return
	}
	if text, ok := b.passage(refs); ok {
		fmt.Println(wrap(text, 78, "  "))
	}
}

func wrap(s string, width int, indent string) string {
	var b strings.Builder
	line := indent
	for _, w := range strings.Fields(s) {
		if len(line)+1+len(w) > width && line != indent {
			b.WriteString(line + "\n")
			line = indent
		}
		if line != indent {
			line += " "
		}
		line += w
	}
	return b.String() + line
}

var bookNames = map[string]string{
	"Gen": "Genesis", "Exod": "Exodus", "Lev": "Leviticus", "Num": "Numbers",
	"Deut": "Deuteronomy", "Josh": "Joshua", "Judg": "Judges", "Ruth": "Ruth",
	"1Sam": "1 Samuel", "2Sam": "2 Samuel", "1Kgs": "1 Kings", "2Kgs": "2 Kings",
	"1Chr": "1 Chronicles", "2Chr": "2 Chronicles", "Ezra": "Ezra", "Neh": "Nehemiah",
	"Tob": "Tobit", "Jdt": "Judith", "Esth": "Esther",
	"1Macc": "1 Maccabees", "2Macc": "2 Maccabees",
	"Job": "Job", "Ps": "Psalms", "Prov": "Proverbs", "Eccl": "Ecclesiastes",
	"Cant": "Song of Songs", "Wis": "Wisdom", "Sir": "Sirach",
	"Isa": "Isaiah", "Jer": "Jeremiah", "Lam": "Lamentations", "Bar": "Baruch",
	"Ezek": "Ezekiel", "Dan": "Daniel", "Hos": "Hosea", "Joel": "Joel",
	"Amos": "Amos", "Obad": "Obadiah", "Jonah": "Jonah", "Mic": "Micah",
	"Nah": "Nahum", "Hab": "Habakkuk", "Zeph": "Zephaniah", "Hag": "Haggai",
	"Zech": "Zechariah", "Mal": "Malachi",
	"Matt": "Matthew", "Mark": "Mark", "Luke": "Luke", "John": "John",
	"Acts": "Acts", "Rom": "Romans", "1Cor": "1 Corinthians", "2Cor": "2 Corinthians",
	"Gal": "Galatians", "Eph": "Ephesians", "Phil": "Philippians", "Col": "Colossians",
	"1Thess": "1 Thessalonians", "2Thess": "2 Thessalonians",
	"1Tim": "1 Timothy", "2Tim": "2 Timothy", "Titus": "Titus", "Phlm": "Philemon",
	"Heb": "Hebrews", "Jas": "James", "1Pet": "1 Peter", "2Pet": "2 Peter",
	"1John": "1 John", "2John": "2 John", "3John": "3 John",
	"Jude": "Jude", "Rev": "Revelation",
}
