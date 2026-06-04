package main

import "testing"

func TestTopDaysFromMap_ReturnsTopThreeSorted(t *testing.T) {
	input := map[string]int{
		"Sunday":    3,
		"Monday":    10,
		"Friday":    4,
		"Wednesday": 10,
		"Tuesday":   8,
	}

	got := topDaysFromMap(input)
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}

	if got[0].Day != "Monday" || got[0].Count != 10 {
		t.Fatalf("unexpected first item: %+v", got[0])
	}
	if got[1].Day != "Wednesday" || got[1].Count != 10 {
		t.Fatalf("unexpected second item: %+v", got[1])
	}
	if got[2].Day != "Tuesday" || got[2].Count != 8 {
		t.Fatalf("unexpected third item: %+v", got[2])
	}
}

func TestFormatTopDaysInline_Empty(t *testing.T) {
	got := formatTopDaysInline(nil)
	if got != "n/a" {
		t.Fatalf("expected n/a, got %q", got)
	}
}

func TestClampMin(t *testing.T) {
	if got := clampMin(2, 5); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
	if got := clampMin(8, 5); got != 8 {
		t.Fatalf("expected 8, got %d", got)
	}
}
