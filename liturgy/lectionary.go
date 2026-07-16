package liturgy

import (
	_ "embed"
	"encoding/json"
)

// Readings are scripture citations for one day. Key formats (see Compute):
//
//	OT-<week>-<weekday 0=Min>-<I|II|A|B|C>  Ordinary Time
//	ADV-<week>-<weekday>  /  ADV-12-<day>   Advent (Dec 17-24 by date)
//	LENT-<week>-<weekday>                   Lent (week 0 = after Rabu Abu)
//	EASTER-<week>-<weekday>                 Easter season
//	XMAS-<mm>-<dd>, BAPTISM-<A|B|C>         Christmas season
//	FEAST-<mm>-<dd>                         fixed solemnities/feasts
type Readings struct {
	First    string `json:"first"`
	Psalm    string `json:"psalm"`
	Second   string `json:"second,omitempty"`   // Sundays & solemnities
	Alleluia string `json:"alleluia,omitempty"` // gospel acclamation verse
	Gospel   string `json:"gospel"`
}

//go:embed lectionary.json
var lectionaryJSON []byte

var lectionary = func() map[string]Readings {
	m := map[string]Readings{}
	if err := json.Unmarshal(lectionaryJSON, &m); err != nil {
		panic("liturgy: bad lectionary.json: " + err.Error())
	}
	return m
}()

// Lookup returns the readings for a computed day, if the table has them.
func (d Day) Lookup() (Readings, bool) {
	r, ok := lectionary[d.Key]
	return r, ok
}
