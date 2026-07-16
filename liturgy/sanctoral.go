package liturgy

import (
	_ "embed"
	"encoding/json"
)

// Saint is a fixed-date celebration from the sanctoral calendar. This is a
// name layer only: memorials keep the weekday readings, so it never affects
// Lookup. Rank is "memorial", "optional memorial", or "feast".
type Saint struct {
	Name string `json:"name"`
	Rank string `json:"rank"`
}

//go:embed sanctoral.json
var sanctoralJSON []byte

var sanctoral = func() map[string][]Saint {
	m := map[string][]Saint{}
	if err := json.Unmarshal(sanctoralJSON, &m); err != nil {
		panic("liturgy: bad sanctoral.json: " + err.Error())
	}
	return m
}()
