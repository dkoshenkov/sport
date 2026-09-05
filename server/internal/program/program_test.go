package program

import (
	"testing"

	"sport/server/internal/api"
)

func TestCalculateUsesXLSXCompatibleFourPercentProgression(t *testing.T) {
	settings := DefaultSettings()
	plan, err := NewCalculator(testProgramOptions()).Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek1})
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
	plan, err := NewCalculator(testProgramOptions()).Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek6})
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
	plan, err := NewCalculator(testProgramOptions()).Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek8})
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
		if _, ok := day.Rows[0].Prescription.Sets.Get(); ok {
			t.Fatalf("%s test row has prescribed sets", day.ID)
		}
	}
}

func TestCalculateMatchesClientXLSXRows(t *testing.T) {
	settings := DefaultSettings()
	settings.Assistance.Deadlift = "paused_deadlift"
	settings.Gpp.Abs = api.NewNilString("abs")
	plan, err := NewCalculator(testProgramOptions()).Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek3})
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
	plan, err := NewCalculator(testProgramOptions()).Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek7})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	assertRow(t, plan.Days[0].Rows[0], api.TrainingRowKindMain, "deadlift", "2x2", 205, "")
	assertRow(t, plan.Days[0].Rows[1], api.TrainingRowKindLight, "bench_press", "3x3", 82.5, "")
	assertRow(t, plan.Days[0].Rows[3], api.TrainingRowKindGpp, "barbell_row", "2x5-6", 0, "RPE: 6")
}

func TestCalculateIncludesPrescribedSetCount(t *testing.T) {
	plan, err := NewCalculator(testProgramOptions()).Calculate(api.ProgramSelection{Settings: DefaultSettings(), Week: api.ProgramWeekWeek3})
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	for _, day := range plan.Days {
		for _, row := range day.Rows {
			if _, ok := row.Prescription.Sets.Get(); !ok {
				t.Fatalf("%s has no prescribed set count", row.RowId)
			}
		}
	}
}

func TestPrescribedSetsParsesTrackablePrescription(t *testing.T) {
	for _, test := range []struct {
		name string
		text string
		want int
		ok   bool
	}{
		{name: "ascii separator", text: "5x5", want: 5, ok: true},
		{name: "cyrillic separator", text: "2х5-6", want: 2, ok: true},
		{name: "test prescription", text: "Тест", ok: false},
		{name: "invalid set count", text: "0x5", ok: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := prescribedSets(test.text).Get()
			if ok != test.ok || (ok && got != test.want) {
				t.Fatalf("prescribedSets(%q) = %d, %v; want %d, %v", test.text, got, ok, test.want, test.ok)
			}
		})
	}
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

func TestCalculatorUsesProvidedExerciseOptions(t *testing.T) {
	options := testProgramOptions()
	options.Assistance.Deadlift = []api.SelectOption{{ID: "custom_deadlift", Label: "Новая тяга"}}
	calculator := NewCalculator(options)
	settings := DefaultSettings()
	settings.Assistance.Deadlift = "custom_deadlift"
	plan, err := calculator.Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek1})
	if err != nil {
		t.Fatal(err)
	}
	if got := plan.Days[0].Rows[2].ExerciseName; got != "Новая тяга" {
		t.Fatalf("name = %q", got)
	}
	settings.Assistance.Bench = "custom_deadlift"
	if _, err = calculator.Calculate(api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek1}); err == nil {
		t.Fatal("accepted exercise from the wrong group")
	}
}

// Minimal inputs for these unit tests; production options are seeded by SQL.
func testProgramOptions() *api.ProgramOptionsResponse {
	return &api.ProgramOptionsResponse{
		Assistance: api.AssistanceOptions{
			Deadlift: []api.SelectOption{{ID: "good_morning", Label: "Гуд-морнинг"}, {ID: "paused_deadlift", Label: "Становая тяга с паузами"}},
			Bench:    []api.SelectOption{{ID: "close_grip_bench", Label: "Жим узким хватом"}},
			Squat:    []api.SelectOption{{ID: "front_squat", Label: "Фронтальный присед"}},
		},
		Gpp: api.GPPOptions{
			Abs:            []api.SelectOption{{ID: "abs", Label: "Пресс"}},
			Triceps:        []api.SelectOption{{ID: "triceps", Label: "Трицепс"}},
			HorizontalPull: []api.SelectOption{{ID: "barbell_row", Label: "Тяга штанги"}},
			Biceps:         []api.SelectOption{{ID: "biceps", Label: "Бицепс"}},
			VerticalPull:   []api.SelectOption{{ID: "pull_up", Label: "Подтягивания"}},
			OverheadPress:  []api.SelectOption{{ID: "kettlebell_military_press", Label: "Жим над головой"}},
		},
	}
}
