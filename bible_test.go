package main

import (
	"encoding/json"
	"testing"
)

func TestConvertAYT(t *testing.T) {
	raw := []byte(`[
		{"book":"1","chapter":"1","verse":"1","text":"<t />Pada mulanya, Allah menciptakan langit dan bumi."},
		{"book":"1","chapter":"1","verse":"2","text":"Bumi tidak berbentuk."},
		{"book":"66","chapter":"22","verse":"21","text":"Amin."}
	]`)
	out, err := convertAYT(raw)
	if err != nil {
		t.Fatal(err)
	}
	var b bibleFile
	if err := json.Unmarshal(out, &b); err != nil {
		t.Fatal(err)
	}
	if len(b.Books) != 2 || b.Books[0].Name != "Genesis" || b.Books[1].Name != "Revelation" {
		t.Fatalf("books = %+v", b.Books)
	}
	v := b.Books[0].Chapters[0].Verses[0]
	if v.Verse != 1 || v.Text != "Pada mulanya, Allah menciptakan langit dan bumi." {
		t.Errorf("verse = %+v (markup not stripped?)", v)
	}
}
