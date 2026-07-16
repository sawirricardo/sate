package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sawirricardo/sate/liturgy"
)

func main() {
	now := time.Now()
	if len(os.Args) > 1 { // sate 2026-12-25 — show the banner for any date
		if t, err := time.Parse("2006-01-02", os.Args[1]); err == nil {
			now = t
		}
	}
	day := liturgy.Compute(now)
	fmt.Printf("✝ %s — %s (Year %s/%s)\n",
		now.Format("Monday, 2 January 2006"), day.Name, day.SundayCycle, day.WeekdayCycle)
	if r, ok := day.Lookup(); ok {
		fmt.Printf("  Readings: %s · %s", r.First, r.Psalm)
		if r.Second != "" {
			fmt.Printf(" · %s", r.Second)
		}
		fmt.Printf(" · %s\n", r.Gospel)
	}
}
