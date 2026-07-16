package main

import "testing"

func TestVulgatePsalm(t *testing.T) {
	for _, tc := range [][2]int{
		{1, 1}, {8, 8}, {9, 9}, {10, 9}, {23, 22}, {51, 50}, {102, 101},
		{113, 112}, {114, 113}, {115, 113}, {116, 114}, {117, 116},
		{130, 129}, {146, 145}, {147, 146}, {148, 148}, {150, 150},
	} {
		if got := vulgatePsalm(tc[0]); got != tc[1] {
			t.Errorf("vulgatePsalm(%d) = %d, want %d", tc[0], got, tc[1])
		}
	}
}

func TestWrap(t *testing.T) {
	got := wrap("aa bb cc dd", 8, "  ")
	if got != "  aa bb\n  cc dd" {
		t.Errorf("wrap = %q", got)
	}
}
