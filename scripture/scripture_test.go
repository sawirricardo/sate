package scripture

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	eoc := EndOfChapter
	for _, tc := range []struct {
		in   string
		want []Ref
	}{
		{"Matt 11:28-30", []Ref{{"Matt", 11, 28, 30}}},
		{"Isa 26:7-9, 12, 16-19", []Ref{{"Isa", 26, 7, 9}, {"Isa", 26, 12, 12}, {"Isa", 26, 16, 19}}},
		{"Ps 102:13-14ab+15, 16-18", []Ref{{"Ps", 102, 13, 14}, {"Ps", 102, 15, 15}, {"Ps", 102, 16, 18}}},
		{"Rev 11:19a; 12:1-6a; 10ab", []Ref{{"Rev", 11, 19, 19}, {"Rev", 12, 1, 6}, {"Rev", 12, 10, 10}}},
		{"2 Cor 5:20-6:2", []Ref{{"2Cor", 5, 20, eoc}, {"2Cor", 6, 1, 2}}},
		{"Matt 26:14-27:66", []Ref{{"Matt", 26, 14, eoc}, {"Matt", 27, 1, 66}}},
		{"3 John 5-8", []Ref{{"3John", 1, 5, 8}}},
		{"Wis 3:1-9 or Wis 4:7-15", []Ref{{"Wis", 3, 1, 9}}},
		{"Joel 2:12-18 and 2 Cor 5:20-6:2", []Ref{{"Joel", 2, 12, 18}, {"2Cor", 5, 20, eoc}, {"2Cor", 6, 1, 2}}},
		{"Luke 24:13-35 ( for afternoon Masses )", []Ref{{"Luke", 24, 13, 35}}},
		{"Eccl/Qoh 1:2-11", []Ref{{"Eccl", 1, 2, 11}}},
		{"Ps 23", []Ref{{"Ps", 23, 1, eoc}}},
		{"Isa 8:23b-9:3", []Ref{{"Isa", 8, 23, eoc}, {"Isa", 9, 1, 3}}},
	} {
		got, err := Parse(tc.in)
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Parse(%q) = %v, %v; want %v", tc.in, got, err, tc.want)
		}
	}
	if _, err := Parse("Esth C:12, 14-16"); err == nil {
		t.Error("Greek Esther chapter should be an error")
	}
}

func TestPickCycle(t *testing.T) {
	bundled := "A: Matt 17:1-9 B: Mark 9:2-10 C: Luke 9:28b-36"
	if got := PickCycle(bundled, "B"); got != "Mark 9:2-10" {
		t.Errorf("PickCycle B = %q", got)
	}
	if got := PickCycle("Matt 11:28-30", "A"); got != "Matt 11:28-30" {
		t.Errorf("PickCycle passthrough = %q", got)
	}
}

// Every citation in the shipped lectionary must parse (per cycle where
// bundled), except the known Greek-Esther day.
func TestCorpus(t *testing.T) {
	raw, err := os.ReadFile("../liturgy/lectionary.json")
	if err != nil {
		t.Fatal(err)
	}
	var table map[string]map[string]string
	if err := json.Unmarshal(raw, &table); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{"LENT-1-4/first": true} // Esth C:12 (Greek addition)
	n := 0
	for key, readings := range table {
		for field, citation := range readings {
			for _, cy := range []string{"A", "B", "C"} {
				if _, err := Parse(PickCycle(citation, cy)); err != nil {
					if !allowed[key+"/"+field] {
						t.Errorf("%s/%s (cycle %s): %v", key, field, cy, err)
					}
					break
				}
			}
			n++
		}
	}
	t.Logf("parsed %d citations", n)
}
