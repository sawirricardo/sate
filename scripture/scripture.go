// Package scripture parses lectionary citation strings like
// "Isa 26:7-9, 12, 16-19" or "Rev 11:19a; 12:1-6a; 10ab" into verse ranges.
package scripture

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Ref is one contiguous verse range within a chapter. Sub-verse letters
// ("14ab") are dropped: text lookup works on whole verses.
type Ref struct {
	Book     string // canonical code: "Gen", "1Sam", "Ps", ...
	Chapter  int
	From, To int // To == EndOfChapter means "through the end of the chapter"
}

// EndOfChapter marks an open range; callers clamp it to the chapter's
// actual verse count.
const EndOfChapter = 1 << 30

// PickCycle resolves citations that bundle all three Sunday cycles
// ("A: Matt 17:1-9 B: Mark 9:2-10 C: Luke 9:28b-36") to one cycle's part.
// Citations without cycle markers are returned unchanged.
var cycleRe = regexp.MustCompile(`(^| )([ABC]): `)

func PickCycle(citation, cycle string) string {
	locs := cycleRe.FindAllStringSubmatchIndex(citation, -1)
	for i, loc := range locs {
		if citation[loc[4]:loc[5]] == cycle {
			end := len(citation)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			return strings.TrimSpace(citation[loc[1]:end])
		}
	}
	return citation
}

// Parse converts one citation to verse ranges. Alternatives ("X or Y") take
// the first; conjunctions ("X and Y") concatenate. An error means the
// citation cannot be resolved to whole verses (e.g. Greek Esther "C:12").
func Parse(citation string) ([]Ref, error) {
	s := regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(citation, " ")
	s = strings.Join(strings.Fields(s), " ")
	s = orRe.Split(s, 2)[0] // alternatives, incl. "or, in Year A, ..."
	s = strings.ReplaceAll(s, " - ", "-")
	var refs []Ref
	for _, part := range strings.Split(s, " and ") {
		r, err := parsePart(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("%q: %w", citation, err)
		}
		refs = append(refs, r...)
	}
	return refs, nil
}

var orRe = regexp.MustCompile(`\s+or\b`)
var bookRe = regexp.MustCompile(`^([123]? ?[A-Za-z]+(?:/[A-Za-z]+)?)\.? (.+)$`)
var crossRe = regexp.MustCompile(`^(\d+):(\d+)[a-z]*-(\d+):(\d+)[a-z]*$`)
var intoChapRe = regexp.MustCompile(`^(\d+)[a-z]*-(\d+):(\d+)[a-z]*$`) // "25-14:1"
var chapRe = regexp.MustCompile(`^(\d+):(.*)$`)
var verseRe = regexp.MustCompile(`^(\d+)[a-z]*$`)
var rangeRe = regexp.MustCompile(`^(\d+)[a-z]*-(\d+)[a-z]*$`)
var lettersRe = regexp.MustCompile(`^[a-z]+$`) // "17b+a": second item is same verse

func parsePart(part string) ([]Ref, error) {
	m := bookRe.FindStringSubmatch(part)
	if m == nil {
		return nil, fmt.Errorf("no book name")
	}
	book, ok := books[normalize(m[1])]
	if !ok {
		return nil, fmt.Errorf("unknown book %q", m[1])
	}
	var refs []Ref
	chapter := 0
	for _, seg := range strings.Split(m[2], ";") {
		seg = strings.TrimSpace(seg)
		// a segment may switch books: "1 Sam 3:9; John 6:68c"
		if bm := bookRe.FindStringSubmatch(seg); bm != nil {
			if b2, ok := books[normalize(bm[1])]; ok {
				book, chapter, seg = b2, 0, bm[2]
			}
		}
		if cm := crossRe.FindStringSubmatch(seg); cm != nil {
			c1, v1, c2, v2 := atoi(cm[1]), atoi(cm[2]), atoi(cm[3]), atoi(cm[4])
			refs = append(refs, Ref{book, c1, v1, EndOfChapter})
			for c := c1 + 1; c < c2; c++ {
				refs = append(refs, Ref{book, c, 1, EndOfChapter})
			}
			refs = append(refs, Ref{book, c2, 1, v2})
			chapter = c2
			continue
		}
		vlist := seg
		if cm := chapRe.FindStringSubmatch(seg); cm != nil {
			chapter = atoi(cm[1])
			vlist = cm[2]
		} else if chapter == 0 {
			if singleChapter[book] {
				chapter = 1 // "3 John 5-8": numbers are verses
			} else {
				// "Ps 23": bare numbers are whole chapters
				for _, item := range splitItems(seg) {
					c, err := wholeChapters(book, item)
					if err != nil {
						return nil, err
					}
					refs = append(refs, c...)
				}
				continue
			}
		}
		for _, item := range splitItems(vlist) {
			switch {
			case lettersRe.MatchString(item): // "17b+a": already covered by previous item
			case intoChapRe.MatchString(item): // "25-14:1": run into the next chapter
				cm := intoChapRe.FindStringSubmatch(item)
				v1, c2, v2 := atoi(cm[1]), atoi(cm[2]), atoi(cm[3])
				refs = append(refs, Ref{book, chapter, v1, EndOfChapter})
				for c := chapter + 1; c < c2; c++ {
					refs = append(refs, Ref{book, c, 1, EndOfChapter})
				}
				refs = append(refs, Ref{book, c2, 1, v2})
				chapter = c2
			case verseRe.MatchString(item):
				v := atoi(verseRe.FindStringSubmatch(item)[1])
				refs = append(refs, Ref{book, chapter, v, v})
			case rangeRe.MatchString(item):
				rm := rangeRe.FindStringSubmatch(item)
				refs = append(refs, Ref{book, chapter, atoi(rm[1]), atoi(rm[2])})
			case crossRe.MatchString(item):
				cm := crossRe.FindStringSubmatch(item)
				c1, v1, c2, v2 := atoi(cm[1]), atoi(cm[2]), atoi(cm[3]), atoi(cm[4])
				refs = append(refs, Ref{book, c1, v1, EndOfChapter})
				for c := c1 + 1; c < c2; c++ {
					refs = append(refs, Ref{book, c, 1, EndOfChapter})
				}
				refs = append(refs, Ref{book, c2, 1, v2})
				chapter = c2
			default:
				return nil, fmt.Errorf("bad verse item %q", item)
			}
		}
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf("no verses")
	}
	return refs, nil
}

func wholeChapters(book, item string) ([]Ref, error) {
	if verseRe.MatchString(item) {
		return []Ref{{book, atoi(verseRe.FindStringSubmatch(item)[1]), 1, EndOfChapter}}, nil
	}
	if rm := rangeRe.FindStringSubmatch(item); rm != nil {
		var refs []Ref
		for c := atoi(rm[1]); c <= atoi(rm[2]); c++ {
			refs = append(refs, Ref{book, c, 1, EndOfChapter})
		}
		return refs, nil
	}
	return nil, fmt.Errorf("bad chapter item %q", item)
}

func splitItems(s string) []string {
	var out []string
	for _, item := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '+' }) {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

func normalize(book string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSuffix(book, "."), " ", ""))
}

var singleChapter = map[string]bool{"Obad": true, "Phlm": true, "2John": true, "3John": true, "Jude": true}

// books maps every abbreviation appearing in the lectionary tables (plus
// common variants) to a canonical code.
var books = map[string]string{
	"gen": "Gen", "exod": "Exod", "lev": "Lev", "num": "Num", "deut": "Deut",
	"josh": "Josh", "judg": "Judg", "ruth": "Ruth",
	"1sam": "1Sam", "2sam": "2Sam", "1kgs": "1Kgs", "2kgs": "2Kgs",
	"1chr": "1Chr", "2chr": "2Chr", "ezra": "Ezra", "neh": "Neh",
	"tob": "Tob", "jdt": "Jdt", "esth": "Esth",
	"1macc": "1Macc", "2macc": "2Macc",
	"job": "Job", "ps": "Ps", "psalm": "Ps", "prov": "Prov",
	"eccl": "Eccl", "eccl/qoh": "Eccl", "cant": "Cant", "song": "Cant",
	"wis": "Wis", "sir": "Sir",
	"isa": "Isa", "jer": "Jer", "lam": "Lam", "bar": "Bar",
	"ezek": "Ezek", "dan": "Dan", "hos": "Hos", "joel": "Joel", "amos": "Amos",
	"obad": "Obad", "jon": "Jonah", "jonah": "Jonah", "mic": "Mic",
	"nah": "Nah", "hab": "Hab", "zeph": "Zeph", "hag": "Hag",
	"zech": "Zech", "mal": "Mal",
	"matt": "Matt", "mark": "Mark", "luke": "Luke", "john": "John",
	"acts": "Acts", "act": "Acts", "rom": "Rom",
	"1cor": "1Cor", "2cor": "2Cor", "gal": "Gal", "eph": "Eph",
	"phil": "Phil", "col": "Col",
	"1thess": "1Thess", "2thess": "2Thess", "1tim": "1Tim", "2tim": "2Tim",
	"titus": "Titus", "phlm": "Phlm", "heb": "Heb", "hebr": "Heb",
	"jas": "Jas", "james": "Jas",
	"1pet": "1Pet", "1peter": "1Pet", "2pet": "2Pet", "2peter": "2Pet",
	"1john": "1John", "2john": "2John", "3john": "3John",
	"jude": "Jude", "rev": "Rev",
}
