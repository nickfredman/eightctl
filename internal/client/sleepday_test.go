package client

import "testing"

func TestSleepDayStageDuration(t *testing.T) {
	day := SleepDay{Stages: []Stage{
		{Stage: "deep", Duration: 3600},
		{Stage: "REM", Duration: 5400},
		{Stage: "light", Duration: 7200},
		{Stage: "rem", Duration: 600},
	}}

	if got := day.DeepDuration(); got != 3600 {
		t.Fatalf("DeepDuration()=%v want 3600", got)
	}
	if got := day.REMDuration(); got != 6000 {
		t.Fatalf("REMDuration()=%v want 6000", got)
	}
	if got := day.StageDuration("deep", "rem"); got != 9600 {
		t.Fatalf("StageDuration(deep, rem)=%v want 9600", got)
	}
}

func TestSleepDayStageDuration_Empty(t *testing.T) {
	day := SleepDay{}
	if got := day.DeepDuration(); got != 0 {
		t.Fatalf("DeepDuration()=%v want 0", got)
	}
	if got := day.REMDuration(); got != 0 {
		t.Fatalf("REMDuration()=%v want 0", got)
	}
}

func TestSleepDayUsesDirectDurationsAndMainSessionFallback(t *testing.T) {
	day := SleepDay{
		SleepDuration: 28800,
		DeepDurationS: 4200,
		REMDurationS:  6000,
	}
	if got := day.DurationSeconds(); got != 28800 {
		t.Fatalf("DurationSeconds()=%v want 28800", got)
	}
	if got := day.DeepDuration(); got != 4200 {
		t.Fatalf("DeepDuration()=%v want 4200", got)
	}
	if got := day.REMDuration(); got != 6000 {
		t.Fatalf("REMDuration()=%v want 6000", got)
	}

	fallback := SleepDay{MainSession: struct {
		Stages []Stage `json:"stages"`
	}{Stages: []Stage{{Stage: "deep", Duration: 1200}, {Stage: "rem", Duration: 900}}}}
	if got := fallback.DeepDuration(); got != 1200 {
		t.Fatalf("fallback DeepDuration()=%v want 1200", got)
	}
	if got := fallback.REMDuration(); got != 900 {
		t.Fatalf("fallback REMDuration()=%v want 900", got)
	}
}
