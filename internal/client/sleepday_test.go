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
