// Package liturgy computes the Roman Catholic liturgical day for a given
// date, offline: computus (Easter), seasons, week numbering, reading cycles.
package liturgy

import (
	"fmt"
	"time"
)

type Season string

const (
	Advent    Season = "Advent"
	Christmas Season = "Christmas"
	Ordinary  Season = "Ordinary Time"
	Lent      Season = "Lent"
	EasterTid Season = "Easter"
)

type Day struct {
	Date         time.Time
	Season       Season
	Week         int    // week within the season (0 = days after Ash Wednesday)
	SundayCycle  string // A, B, C
	WeekdayCycle string // I, II
	Name         string // e.g. "Thursday of the 15th week in Ordinary Time"
	Key          string // lectionary lookup key, e.g. "OT-15-4-II"
}

// Easter returns Easter Sunday for a year (Anonymous Gregorian / Meeus algorithm).
func Easter(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return date(year, time.Month(month), day)
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Advent1 returns the First Sunday of Advent (4th Sunday before Christmas).
func Advent1(year int) time.Time {
	christmas := date(year, time.December, 25)
	advent4 := christmas.AddDate(0, 0, -int(christmas.Weekday()))
	if christmas.Weekday() == time.Sunday {
		advent4 = advent4.AddDate(0, 0, -7)
	}
	return advent4.AddDate(0, 0, -21)
}

// Epiphany is the Sunday between Jan 2-8 (transferred, as in Indonesia).
func Epiphany(year int) time.Time {
	jan2 := date(year, time.January, 2)
	return jan2.AddDate(0, 0, (7-int(jan2.Weekday()))%7)
}

// BaptismOfTheLord is the Sunday after Epiphany, or the Monday right after
// when Epiphany lands on Jan 7/8.
func BaptismOfTheLord(year int) time.Time {
	epiphany := Epiphany(year)
	if epiphany.Day() >= 7 {
		return epiphany.AddDate(0, 0, 1)
	}
	return epiphany.AddDate(0, 0, 7)
}

// Compute returns the liturgical day for a calendar date.
func Compute(t time.Time) Day {
	d := date(t.Year(), t.Month(), t.Day())
	y := d.Year()

	easter := Easter(y)
	ashWed := easter.AddDate(0, 0, -46)
	pentecost := easter.AddDate(0, 0, 49)
	advent1 := Advent1(y)
	baptism := BaptismOfTheLord(y)

	litYear := y
	if !d.Before(advent1) {
		litYear = y + 1
	}
	day := Day{
		Date:         d,
		SundayCycle:  [...]string{"C", "A", "B"}[litYear%3],
		WeekdayCycle: [...]string{"II", "I"}[litYear%2],
	}
	wd := int(d.Weekday())
	wdName := d.Weekday().String()
	sundayOfWeek := d.AddDate(0, 0, -wd) // Sunday starting this week
	weeksBetween := func(from, to time.Time) int { return int(to.Sub(from).Hours()) / (24 * 7) }

	switch {
	case !d.Before(advent1) && d.Before(date(y, time.December, 25)):
		day.Season = Advent
		day.Week = weeksBetween(advent1, sundayOfWeek) + 1
		day.Key = fmt.Sprintf("ADV-%d-%d", day.Week, wd)
		day.Name = fmt.Sprintf("%s of the %s week of Advent", wdName, ordinal(day.Week))
		switch {
		case wd == 0: // Sunday readings follow the A/B/C cycle
			day.Key += "-" + day.SundayCycle
		case d.Month() == time.December && d.Day() >= 17: // weekdays Dec 17-24: proper readings per date
			day.Key = fmt.Sprintf("ADV-12-%02d", d.Day())
		}
	case !d.Before(date(y, time.December, 25)) || !d.After(baptism):
		day.Season = Christmas
		epiphany := Epiphany(y)
		switch {
		case d.Equal(baptism):
			day.Name = "The Baptism of the Lord"
			day.Key = "BAPTISM-" + day.SundayCycle
		case d.Equal(epiphany):
			day.Name = "The Epiphany of the Lord"
			day.Key = "EPIPH"
		case d.After(epiphany) && d.Month() == time.January:
			day.Name = fmt.Sprintf("%s after Epiphany", wdName)
			day.Key = fmt.Sprintf("EPIPH-%d", wd)
		case wd == 0 && d.Month() == time.December:
			day.Name = "The Holy Family"
			day.Key = "HOLYFAM-" + day.SundayCycle
		default:
			day.Name = fmt.Sprintf("%s, Christmas Time", wdName)
			day.Key = fmt.Sprintf("XMAS-%02d-%02d", d.Month(), d.Day())
		}
	case d.Before(ashWed):
		day.Season = Ordinary
		// ponytail: if Baptism falls on a Monday, its week counts as week 1
		day.Week = weeksBetween(baptism.AddDate(0, 0, -int(baptism.Weekday())), sundayOfWeek) + 1
		fillOrdinary(&day, wd, wdName)
	case d.Before(easter):
		day.Season = Lent
		lent1 := ashWed.AddDate(0, 0, 4) // First Sunday of Lent
		if d.Before(lent1) {
			day.Week = 0
			day.Key = fmt.Sprintf("LENT-0-%d", wd)
			day.Name = wdName + " after Ash Wednesday"
			if d.Equal(ashWed) {
				day.Name = "Ash Wednesday"
			}
		} else {
			day.Week = weeksBetween(lent1, sundayOfWeek) + 1
			day.Key = fmt.Sprintf("LENT-%d-%d", day.Week, wd)
			if wd == 0 {
				day.Key += "-" + day.SundayCycle
			}
			day.Name = fmt.Sprintf("%s of the %s week of Lent", wdName, ordinal(day.Week))
			switch {
			case d.Equal(easter.AddDate(0, 0, -7)):
				day.Name = "Palm Sunday"
			case day.Week == 6:
				day.Name = wdName + " of Holy Week"
				if wd >= 4 { // Triduum
					day.Name = [...]string{"", "", "", "", "Holy Thursday", "Good Friday", "Holy Saturday"}[wd]
				}
			}
		}
	case !d.After(pentecost):
		day.Season = EasterTid
		day.Week = weeksBetween(easter, sundayOfWeek) + 1
		day.Key = fmt.Sprintf("EASTER-%d-%d", day.Week, wd)
		if wd == 0 {
			day.Key += "-" + day.SundayCycle
		}
		day.Name = fmt.Sprintf("%s of the %s week of Easter", wdName, ordinal(day.Week))
		switch {
		case d.Equal(easter):
			day.Name = "Easter Sunday"
		case day.Week == 1:
			day.Name = wdName + " within the Octave of Easter"
		case d.Equal(easter.AddDate(0, 0, 39)):
			day.Name = "The Ascension of the Lord" // ponytail: Indonesia keeps Thursday
			day.Key = "ASCENSION-" + day.SundayCycle
		case d.Equal(pentecost):
			day.Name = "Pentecost Sunday"
			day.Key = "PENTECOST-" + day.SundayCycle
		}
	default:
		day.Season = Ordinary
		christKing := advent1.AddDate(0, 0, -7) // Sunday of week 34
		day.Week = 34 - weeksBetween(sundayOfWeek, christKing)
		fillOrdinary(&day, wd, wdName)
		switch {
		case d.Equal(pentecost.AddDate(0, 0, 7)):
			day.Name = "Trinity Sunday"
			day.Key = "TRINITY-" + day.SundayCycle
		case d.Equal(easter.AddDate(0, 0, 63)):
			day.Name = "Corpus Christi" // ponytail: transferred to Sunday, as in Indonesia
			day.Key = "CORPUS-" + day.SundayCycle
		case d.Equal(easter.AddDate(0, 0, 68)):
			day.Name = "The Sacred Heart of Jesus"
			day.Key = "SACREDHEART-" + day.SundayCycle
		case d.Equal(christKing):
			day.Name = "Christ the King"
		}
	}

	// Fixed-date solemnities/feasts override the ferial name (not during
	// Triduum/Sundays precedence subtleties — good enough for a banner).
	// ponytail: name override only; add proper readings keys when the
	// lectionary table gains those entries.
	if f, ok := feasts[d.Format("01-02")]; ok && (day.Season != Lent || f.overridesLent) {
		day.Name = f.name
		day.Key = "FEAST-" + d.Format("01-02")
	}
	return day
}

func fillOrdinary(day *Day, wd int, wdName string) {
	cycle := day.WeekdayCycle
	if wd == 0 {
		cycle = day.SundayCycle
	}
	day.Key = fmt.Sprintf("OT-%d-%d-%s", day.Week, wd, cycle)
	if wd == 0 {
		day.Name = fmt.Sprintf("%s Sunday in Ordinary Time", ordinal(day.Week))
	} else {
		day.Name = fmt.Sprintf("%s of the %s week in Ordinary Time", wdName, ordinal(day.Week))
	}
}

type feast struct {
	name          string
	overridesLent bool
}

var feasts = map[string]feast{
	"01-01": {"Mary, the Holy Mother of God", false},
	"02-02": {"The Presentation of the Lord", false},
	"03-19": {"St. Joseph, Spouse of the Blessed Virgin Mary", true},
	"03-25": {"The Annunciation of the Lord", true},
	"06-24": {"The Nativity of St. John the Baptist", false},
	"06-29": {"Sts. Peter and Paul, Apostles", false},
	"08-06": {"The Transfiguration of the Lord", false},
	"08-15": {"The Assumption of the Blessed Virgin Mary", false},
	"09-14": {"The Exaltation of the Holy Cross", false},
	"11-01": {"All Saints", false},
	"11-02": {"All Souls", false},
	"11-09": {"The Dedication of the Lateran Basilica", false},
	"12-08": {"The Immaculate Conception of the Blessed Virgin Mary", false},
	"12-25": {"The Nativity of the Lord", false},
}

func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}
