package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sawirricardo/sate/liturgy"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "bible" {
		if err := bibleCmd(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "sate:", err)
			os.Exit(1)
		}
		return
	}
	now := time.Now()
	if len(os.Args) > 1 { // sate 2026-12-25 — show the banner for any date
		if t, err := time.Parse("2006-01-02", os.Args[1]); err == nil {
			now = t
		}
	}
	day := liturgy.Compute(now)
	fmt.Printf("✝ %s — %s (Year %s/%s)\n",
		now.Format("Monday, 2 January 2006"), day.Name, day.SundayCycle, day.WeekdayCycle)
	for _, s := range day.Saints {
		fmt.Printf("  %s%s: %s\n", strings.ToUpper(s.Rank[:1]), s.Rank[1:], s.Name)
	}
	if r, ok := day.Lookup(); ok {
		fmt.Printf("  Readings: %s · %s", r.First, r.Psalm)
		if r.Second != "" {
			fmt.Printf(" · %s", r.Second)
		}
		fmt.Printf(" · %s\n", r.Gospel)
		if r.Alleluia != "" {
			label := "Alleluia"
			if day.Season == liturgy.Lent { // no Alleluia during Lent
				label = "Verse before the Gospel"
			}
			fmt.Printf("  %s: %s\n", label, r.Alleluia)
		}
	}
}
