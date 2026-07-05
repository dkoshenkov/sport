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
	for _, day := range plan.Days {
		if len(day.Rows) != 1 {
			t.Fatalf("%s rows = %d, want only test row", day.ID, len(day.Rows))
		}
		if day.Rows[0].Kind != api.TrainingRowKindMain {
			t.Fatalf("%s row kind = %s, want main", day.ID, day.Rows[0].Kind)
		}
	}
}

func TestCalculateMatchesClientXLSXRows(t *testing.T) {
	settings := DefaultSettings()
	settings.Assistance.Deadlift = "paused_deadlift"
	settings.Gpp.Abs = api.NewNilString("abs")
	plan, err := Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek3})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if got, want := plan.Days[0].Focus, "Тяжелая становая"; got != want {
		t.Fatalf("day 1 focus = %q, want %q", got, want)
	}
	if got, want := len(plan.Days[0].Rows), 5; got != want {
		t.Fatalf("day 1 row count = %d, want %d", got, want)
	}

	assertRow(t, plan.Days[0].Rows[1], api.TrainingRowKindLight, "bench_press", "4x8", 75, "")
	assertRow(t, plan.Days[0].Rows[2], api.TrainingRowKindAssistance, "paused_deadlift", "2x4", 0, "RPE: 6-7")
	assertRow(t, plan.Days[1].Rows[1], api.TrainingRowKindLight, "squat", "5x4", 112.5, "")
	assertRow(t, plan.Days[2].Rows[1], api.TrainingRowKindLight, "deadlift", "5x4", 147.5, "")
}

func TestCalculateWeekSevenPatterns(t *testing.T) {
	settings := DefaultSettings()
	plan, err := Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek7})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	assertRow(t, plan.Days[0].Rows[0], api.TrainingRowKindMain, "deadlift", "2x2", 205, "")
	assertRow(t, plan.Days[0].Rows[1], api.TrainingRowKindLight, "bench_press", "3x3", 82.5, "")
	assertRow(t, plan.Days[0].Rows[3], api.TrainingRowKindGpp, "barbell_row", "2x5-6", 0, "RPE: 6")
}

func assertRow(t *testing.T, row api.TrainingRow, kind api.TrainingRowKind, exerciseKey, setsReps string, weight float64, rpe string) {
	t.Helper()
	if row.Kind != kind {
		t.Fatalf("%s kind = %s, want %s", row.RowId, row.Kind, kind)
	}
	if row.ExerciseKey != exerciseKey {
		t.Fatalf("%s exercise key = %s, want %s", row.RowId, row.ExerciseKey, exerciseKey)
	}
	if row.Prescription.SetsRepsText != setsReps {
		t.Fatalf("%s sets/reps = %s, want %s", row.RowId, row.Prescription.SetsRepsText, setsReps)
	}
	if weight > 0 {
		got, ok := row.Prescription.WeightKg.Get()
		if !ok {
			t.Fatalf("%s has no weight, want %v", row.RowId, weight)
		}
		if got != weight {
			t.Fatalf("%s weight = %v, want %v", row.RowId, got, weight)
		}
	}
	if rpe != "" {
		got, ok := row.Prescription.RpeText.Get()
		if !ok || got != rpe {
			t.Fatalf("%s rpe = %q, ok = %v; want %q", row.RowId, got, ok, rpe)
		}
	}
}
