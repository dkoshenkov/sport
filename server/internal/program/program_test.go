package program

import (
	"testing"

	"sport/server/internal/api"
)

func TestCalculateUsesXLSXCompatibleFourPercentProgression(t *testing.T) {
	settings := DefaultSettings()
	plan, err := Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek1})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	got, ok := plan.Days[0].Rows[0].Prescription.WeightKg.Get()
	if !ok {
		t.Fatal("deadlift main row has no weight")
	}
	if got != 145 {
		t.Fatalf("deadlift week 1 weight = %v, want 145", got)
	}
}

func TestCalculateAllowsExplicitFivePercentProgression(t *testing.T) {
	settings := DefaultSettings()
	settings.ProgressionStep = api.ProgressionStepStep5Percent
	plan, err := Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek6})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	got, ok := plan.Days[0].Rows[0].Prescription.WeightKg.Get()
	if !ok {
		t.Fatal("deadlift main row has no weight")
	}
	if got != 197.5 {
		t.Fatalf("deadlift week 6 weight = %v, want 197.5", got)
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected warning for explicit 5% progression")
	}
}

func TestCalculateWeekEightUsesOneRepMaxText(t *testing.T) {
	settings := DefaultSettings()
	plan, err := Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek8})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	weightText, ok := plan.Days[0].Rows[0].Prescription.WeightText.Get()
	if !ok || weightText != "1ПМ" {
		t.Fatalf("week 8 weight text = %q, ok = %v; want 1ПМ", weightText, ok)
	}
	if _, ok := plan.Days[0].Rows[0].Prescription.WeightKg.Get(); ok {
		t.Fatal("week 8 should not set numeric working weight")
	}
}
