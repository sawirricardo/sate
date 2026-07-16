package liturgy

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time { return date(y, m, day) }

func TestEaster(t *testing.T) {
	for _, tc := range []struct {
		year int
		want time.Time
	}{
		{2024, d(2024, time.March, 31)},
		{2025, d(2025, time.April, 20)},
		{2026, d(2026, time.April, 5)},
		{2030, d(2030, time.April, 21)},
	} {
		if got := Easter(tc.year); !got.Equal(tc.want) {
			t.Errorf("Easter(%d) = %v, want %v", tc.year, got, tc.want)
		}
	}
}

func TestAdvent1(t *testing.T) {
	if got := Advent1(2026); !got.Equal(d(2026, time.November, 29)) {
		t.Errorf("Advent1(2026) = %v", got)
	}
	// Christmas 2022 was a Sunday -> Advent 1 on Nov 27
	if got := Advent1(2022); !got.Equal(d(2022, time.November, 27)) {
		t.Errorf("Advent1(2022) = %v", got)
	}
}

func TestCompute(t *testing.T) {
	for _, tc := range []struct {
		date       time.Time
		season     Season
		week       int
		key        string
		sun, wkday string
	}{
		// Thursday, 15th week of Ordinary Time, Year A / II
		{d(2026, time.July, 16), Ordinary, 15, "OT-15-4-II", "A", "II"},
		// Tuesday of OT week 1 (Baptism was Sun Jan 11, 2026)
		{d(2026, time.January, 13), Ordinary, 1, "OT-1-2-II", "A", "II"},
		// Ash Wednesday 2026
		{d(2026, time.February, 18), Lent, 0, "LENT-0-3", "A", "II"},
		// Easter Sunday
		{d(2026, time.April, 5), EasterTid, 1, "EASTER-1-0-A", "A", "II"},
		// Tuesday of Advent week 1, liturgical year 2027 = Year B / I
		{d(2026, time.December, 1), Advent, 1, "ADV-1-2", "B", "I"},
		// Dec 18 uses date-keyed Advent readings
		{d(2026, time.December, 18), Advent, 3, "ADV-12-18", "B", "I"},
	} {
		got := Compute(tc.date)
		if got.Season != tc.season || got.Week != tc.week || got.Key != tc.key ||
			got.SundayCycle != tc.sun || got.WeekdayCycle != tc.wkday {
			t.Errorf("Compute(%s) = %+v, want season=%s week=%d key=%s %s/%s",
				tc.date.Format("2006-01-02"), got, tc.season, tc.week, tc.key, tc.sun, tc.wkday)
		}
	}
}

func TestNames(t *testing.T) {
	for _, tc := range []struct {
		date time.Time
		name string
	}{
		{d(2026, time.July, 16), "Thursday of the 15th week in Ordinary Time"},
		{d(2026, time.April, 3), "Good Friday"},
		{d(2026, time.March, 29), "Palm Sunday"},
		{d(2026, time.May, 14), "The Ascension of the Lord"},
		{d(2026, time.May, 24), "Pentecost Sunday"},
		{d(2026, time.August, 15), "The Assumption of the Blessed Virgin Mary"},
		{d(2026, time.November, 22), "Christ the King"},
		{d(2026, time.January, 11), "The Baptism of the Lord"},
	} {
		if got := Compute(tc.date); got.Name != tc.name {
			t.Errorf("Compute(%s).Name = %q, want %q", tc.date.Format("2006-01-02"), got.Name, tc.name)
		}
	}
}

// Every day from 2024 through 2030 must resolve to a reading, except days
// with no Mass of the day (Holy Saturday) — the banner shows nothing there.
func TestFullCoverage(t *testing.T) {
	allowedMissing := map[string]bool{"LENT-6-6": true}
	missing := map[string][]string{}
	for d := d(2024, time.January, 1); d.Year() < 2031; d = d.AddDate(0, 0, 1) {
		day := Compute(d)
		if _, ok := day.Lookup(); !ok && !allowedMissing[day.Key] {
			missing[day.Key] = append(missing[day.Key], d.Format("2006-01-02"))
		}
	}
	for key, dates := range missing {
		t.Errorf("no readings for %s (e.g. %s, %d days)", key, dates[0], len(dates))
	}
}

func TestSaints(t *testing.T) {
	got := Compute(d(2026, time.July, 16)).Saints
	if len(got) != 1 || got[0].Name != "Our Lady of Mount Carmel" || got[0].Rank != "optional memorial" {
		t.Errorf("Jul 16 Saints = %+v", got)
	}
	// US-only entries must be excluded
	if got := Compute(d(2026, time.January, 4)).Saints; len(got) != 0 {
		t.Errorf("Jan 4 (St. Elizabeth Ann Seton, USA) should be empty, got %+v", got)
	}
	// feast-override days carry no duplicate saint line
	if got := Compute(d(2026, time.August, 15)).Saints; len(got) != 0 {
		t.Errorf("Assumption day should have no saint line, got %+v", got)
	}
}

func TestLookup(t *testing.T) {
	// Verified against USCCB's published readings for this liturgical day.
	r, ok := Compute(d(2026, time.July, 16)).Lookup()
	if !ok || r.Gospel != "Matt 11:28-30" || r.First != "Isa 26:7-9, 12, 16-19" || r.Alleluia != "Matt 11:28" {
		t.Errorf("Lookup = %+v ok=%v", r, ok)
	}
}
