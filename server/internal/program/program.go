package program

import (
	"fmt"
	"math"

	"sport/server/internal/api"
)

const FormulaVersion = "xlsx-linear-cycle-v1"

type ExerciseOption struct {
	ID    string
	Label string
}

func Options() *api.ProgramOptionsResponse {
	return &api.ProgramOptionsResponse{
		Weeks: []api.SelectOption{
			option("week_1", "Неделя 1"), option("week_2", "Неделя 2"),
			option("week_3", "Неделя 3"), option("week_4", "Неделя 4"),
			option("week_5", "Неделя 5"), option("week_6", "Неделя 6"),
			option("week_7", "Неделя 7"), option("week_8", "Неделя 8"),
		},
		Variants: []api.SelectOption{
			option("variant_1", "Вариант 1"),
			option("variant_2", "Вариант 2"),
		},
		ProgressionSteps: []api.SelectOption{
			option("step_4_percent", "4% от 1ПМ"),
			option("step_5_percent", "5% от 1ПМ"),
		},
		Assistance: api.AssistanceOptions{
			Deadlift: toAPIOptions(deadliftAssistance),
			Bench:    toAPIOptions(benchAssistance),
			Squat:    toAPIOptions(squatAssistance),
		},
		Gpp: api.GPPOptions{
			Abs:            toAPIOptions(withSkip(gppAbs)),
			Triceps:        toAPIOptions(withSkip(gppTriceps)),
			HorizontalPull: toAPIOptions(withSkip(gppHorizontalPull)),
			Biceps:         toAPIOptions(withSkip(gppBiceps)),
			VerticalPull:   toAPIOptions(withSkip(gppVerticalPull)),
			OverheadPress:  toAPIOptions(withSkip(gppOverheadPress)),
		},
	}
}

func Calculate(selection api.ProgramSelection) (*api.TrainingPlanResponse, error) {
	settings := normalizeSettings(selection.Settings)
	if err := validateSettings(settings); err != nil {
		return nil, err
	}
	week := selection.Week

	days := []api.TrainingDay{
		{
			ID:    api.TrainingDayIDDay1,
			Label: "День 1",
			Focus: "Тяжелая становая",
			Rows: compactRows([]api.TrainingRow{
				heavyRow("day_1_deadlift", "deadlift", "Становая тяга", settings.OneRepMaxKg.Deadlift, settings, week),
				lightRow("day_1_bench", "bench_press", "Жим лежа", settings.OneRepMaxKg.Bench, week, true),
				assistanceRow("day_1_deadlift_assistance", settings.Assistance.Deadlift, week, settings.Assistance.Deadlift == "paused_deadlift"),
				gppPullRow("day_1_horizontal_pull", settings.Gpp.HorizontalPull, week),
				gppArmRow("day_1_biceps", settings.Gpp.Biceps, week),
			}),
		},
		{
			ID:    api.TrainingDayIDDay2,
			Label: "День 2",
			Focus: "Тяжелый жим",
			Rows: compactRows([]api.TrainingRow{
				heavyRow("day_2_bench", "bench_press", "Жим лежа", settings.OneRepMaxKg.Bench, settings, week),
				lightRow("day_2_squat", "squat", "Приседания", settings.OneRepMaxKg.Squat, week, false),
				assistanceRow("day_2_bench_assistance", settings.Assistance.Bench, week, false),
				gppSimpleRow("day_2_abs", settings.Gpp.Abs, week, "3x6-10", "RPE: 7"),
				gppArmRow("day_2_triceps", settings.Gpp.Triceps, week),
			}),
		},
		{
			ID:    api.TrainingDayIDDay3,
			Label: "День 3",
			Focus: "Тяжелый присед",
			Rows: compactRows([]api.TrainingRow{
				heavyRow("day_3_squat", "squat", "Приседания", settings.OneRepMaxKg.Squat, settings, week),
				lightRow("day_3_deadlift", "deadlift", "Становая тяга", settings.OneRepMaxKg.Deadlift, week, false),
				assistanceRow("day_3_squat_assistance", settings.Assistance.Squat, week, false),
				gppPullRow("day_3_vertical_pull", settings.Gpp.VerticalPull, week),
				gppSimpleRow("day_3_overhead_press", settings.Gpp.OverheadPress, week, "3x6-10", overheadPressRPE(week)),
			}),
		},
	}

	warnings := []string{}
	if settings.ProgressionStep == api.ProgressionStepStep5Percent {
		warnings = append(warnings, "Используется явный шаг прогрессии 5%; XLSX-совместимое значение по умолчанию - 4%.")
	}

	return &api.TrainingPlanResponse{
		Selection: api.ProgramSelection{
			Settings: settings,
			Week:     selection.Week,
		},
		Days:           days,
		FormulaVersion: FormulaVersion,
		Warnings:       warnings,
	}, nil
}

func DefaultSettings() api.CycleSettings {
	return api.CycleSettings{
		OneRepMaxKg:     api.OneRepMaxes{Deadlift: 225, Bench: 125, Squat: 170},
		Variant:         api.ProgramVariantVariant1,
		ProgressionStep: api.ProgressionStepStep4Percent,
		Assistance: api.AssistanceSelection{
			Deadlift: "good_morning",
			Bench:    "close_grip_bench",
			Squat:    "front_squat",
		},
		Gpp: api.GPPSelection{
			Abs:            api.NewNilString("abs"),
			Triceps:        api.NewNilString("triceps"),
			HorizontalPull: api.NewNilString("barbell_row"),
			Biceps:         api.NewNilString("biceps"),
			VerticalPull:   api.NewNilString("pull_up"),
			OverheadPress:  api.NewNilString("kettlebell_military_press"),
		},
	}
}

func roundToNearest2_5(value float64) float64 {
	return math.Round(value/2.5) * 2.5
}

func baseWeight(oneRepMax float64, variant api.ProgramVariant) float64 {
	if variant == api.ProgramVariantVariant2 {
		return oneRepMax * 0.86
	}
	return oneRepMax * 0.82
}

func progressionStep(oneRepMax float64, step api.ProgressionStep) float64 {
	if step == api.ProgressionStepStep5Percent {
		return roundToNearest2_5(oneRepMax * 0.05)
	}
	return roundToNearest2_5(oneRepMax * 0.04)
}

func validateSettings(settings api.CycleSettings) error {
	if settings.OneRepMaxKg.Deadlift <= 0 || settings.OneRepMaxKg.Bench <= 0 || settings.OneRepMaxKg.Squat <= 0 {
		return fmt.Errorf("one-rep max values must be positive")
	}
	for _, key := range []string{settings.Assistance.Deadlift, settings.Assistance.Bench, settings.Assistance.Squat} {
		if _, ok := optionLabels[key]; !ok {
			return fmt.Errorf("unknown exercise option %q", key)
		}
	}
	for _, opt := range []api.NilString{
		settings.Gpp.Abs,
		settings.Gpp.Triceps,
		settings.Gpp.HorizontalPull,
		settings.Gpp.Biceps,
		settings.Gpp.VerticalPull,
		settings.Gpp.OverheadPress,
	} {
		key, ok := opt.Get()
		if !ok || key == "" || key == "-" {
			continue
		}
		if _, exists := optionLabels[key]; !exists {
			return fmt.Errorf("unknown gpp option %q", key)
		}
	}
	return nil
}

func heavyRow(rowID, key, name string, oneRepMax float64, settings api.CycleSettings, week api.ProgramWeek) api.TrainingRow {
	if week == api.ProgramWeekWeek8 {
		return row(rowID, key, name, api.TrainingRowKindMain, api.Prescription{
			SetsRepsText: "Тест",
			WeightText:   api.NewOptNilString("1ПМ"),
		})
	}

	weight := roundToNearest2_5(baseWeight(oneRepMax, settings.Variant) + progressionStep(oneRepMax, settings.ProgressionStep)*float64(weeklyOffset(week)))
	setsReps := "5x5"
	if week == api.ProgramWeekWeek6 {
		setsReps = "3x3"
	}
	if week == api.ProgramWeekWeek7 {
		setsReps = "2x2"
	}
	return row(rowID, key, name, api.TrainingRowKindMain, api.Prescription{
		SetsRepsText: setsReps,
		WeightKg:     api.NewOptNilFloat64(weight),
		Unit:         api.NewOptNilString("т"),
	})
}

func lightRow(rowID, key, name string, oneRepMax float64, week api.ProgramWeek, bench bool) api.TrainingRow {
	if week == api.ProgramWeekWeek8 {
		return api.TrainingRow{}
	}
	percent := lightLiftPercent[week]
	if bench {
		percent = lightBenchPercent[week]
	}
	setsReps := lightLiftPrescription[week]
	if bench {
		switch week {
		case api.ProgramWeekWeek1, api.ProgramWeekWeek3:
			setsReps = "4x8"
		case api.ProgramWeekWeek7:
			setsReps = "3x3"
		default:
			setsReps = "5x5"
		}
	}
	return row(rowID, key, name, api.TrainingRowKindLight, api.Prescription{
		SetsRepsText: setsReps,
		WeightKg:     api.NewOptNilFloat64(roundToNearest2_5(oneRepMax * percent)),
		Unit:         api.NewOptNilString("л"),
	})
}

func assistanceRow(rowID, key string, week api.ProgramWeek, forcePausedDeadliftPattern bool) api.TrainingRow {
	if week == api.ProgramWeekWeek8 || key == "" {
		return api.TrainingRow{}
	}
	setsReps := "2x6"
	rpe := "RPE: 6-7"
	if forcePausedDeadliftPattern {
		setsReps = "2x4"
	} else if week == api.ProgramWeekWeek6 {
		setsReps = "3x4"
		rpe = "RPE: 5"
	} else if week == api.ProgramWeekWeek7 {
		setsReps = "2x3"
		rpe = "RPE: 5"
	}
	return row(rowID, key, label(key), api.TrainingRowKindAssistance, api.Prescription{
		SetsRepsText: setsReps,
		RpeText:      api.NewOptNilString(rpe),
	})
}

func gppPullRow(rowID string, opt api.NilString, week api.ProgramWeek) api.TrainingRow {
	key, ok := opt.Get()
	if week == api.ProgramWeekWeek8 || !ok || key == "" || key == "-" {
		return api.TrainingRow{}
	}
	setsReps := "3x6-10"
	rpe := "RPE: 7"
	if week == api.ProgramWeekWeek7 {
		setsReps = "2x5-6"
		rpe = "RPE: 6"
	}
	return row(rowID, key, label(key), api.TrainingRowKindGpp, api.Prescription{
		SetsRepsText: setsReps,
		RpeText:      api.NewOptNilString(rpe),
	})
}

func gppArmRow(rowID string, opt api.NilString, week api.ProgramWeek) api.TrainingRow {
	key, ok := opt.Get()
	if week == api.ProgramWeekWeek8 || !ok || key == "" || key == "-" {
		return api.TrainingRow{}
	}
	sets := armSetsByWeek[week]
	rpe := "RPE: 8"
	if week == api.ProgramWeekWeek7 {
		rpe = "RPE: 6"
	}
	return row(rowID, key, label(key), api.TrainingRowKindGpp, api.Prescription{
		SetsRepsText: fmt.Sprintf("%dx6-12", sets),
		RpeText:      api.NewOptNilString(rpe),
	})
}

func gppSimpleRow(rowID string, opt api.NilString, week api.ProgramWeek, setsReps, rpe string) api.TrainingRow {
	key, ok := opt.Get()
	if week == api.ProgramWeekWeek8 || !ok || key == "" || key == "-" {
		return api.TrainingRow{}
	}
	return row(rowID, key, label(key), api.TrainingRowKindGpp, api.Prescription{
		SetsRepsText: setsReps,
		RpeText:      api.NewOptNilString(rpe),
	})
}

func compactRows(rows []api.TrainingRow) []api.TrainingRow {
	out := rows[:0]
	for _, row := range rows {
		if row.ExerciseKey != "" {
			out = append(out, row)
		}
	}
	return out
}

func row(rowID, key, name string, kind api.TrainingRowKind, prescription api.Prescription) api.TrainingRow {
	return api.TrainingRow{RowId: rowID, ExerciseKey: key, ExerciseName: name, Kind: kind, Prescription: prescription}
}

func option(id, label string) api.SelectOption {
	return api.SelectOption{ID: id, Label: label}
}

func toAPIOptions(items []ExerciseOption) []api.SelectOption {
	out := make([]api.SelectOption, 0, len(items))
	for _, item := range items {
		out = append(out, option(item.ID, item.Label))
	}
	return out
}

func withSkip(items []ExerciseOption) []ExerciseOption {
	out := make([]ExerciseOption, 0, len(items)+1)
	out = append(out, ExerciseOption{ID: "", Label: "-"})
	out = append(out, items...)
	return out
}

func label(key string) string {
	if v, ok := optionLabels[key]; ok {
		return v
	}
	return key
}

func overheadPressRPE(week api.ProgramWeek) string {
	if week == api.ProgramWeekWeek7 {
		return "RPE: 5"
	}
	return "RPE: 6"
}

func weeklyOffset(week api.ProgramWeek) int {
	switch week {
	case api.ProgramWeekWeek1:
		return -4
	case api.ProgramWeekWeek2:
		return -3
	case api.ProgramWeekWeek3:
		return -2
	case api.ProgramWeekWeek4:
		return -1
	case api.ProgramWeekWeek6:
		return 1
	case api.ProgramWeekWeek7:
		return 2
	default:
		return 0
	}
}

func normalizeSettings(settings api.CycleSettings) api.CycleSettings {
	settings.Assistance.Deadlift = normalizeKey(settings.Assistance.Deadlift)
	settings.Assistance.Bench = normalizeKey(settings.Assistance.Bench)
	settings.Assistance.Squat = normalizeKey(settings.Assistance.Squat)
	settings.Gpp.Abs = normalizeNilString(settings.Gpp.Abs)
	settings.Gpp.Triceps = normalizeNilString(settings.Gpp.Triceps)
	settings.Gpp.HorizontalPull = normalizeNilString(settings.Gpp.HorizontalPull)
	settings.Gpp.Biceps = normalizeNilString(settings.Gpp.Biceps)
	settings.Gpp.VerticalPull = normalizeNilString(settings.Gpp.VerticalPull)
	settings.Gpp.OverheadPress = normalizeNilString(settings.Gpp.OverheadPress)
	return settings
}

func normalizeNilString(value api.NilString) api.NilString {
	key, ok := value.Get()
	if !ok || key == "" || key == "-" {
		var out api.NilString
		out.SetToNull()
		return out
	}
	return api.NewNilString(normalizeKey(key))
}

func normalizeKey(key string) string {
	if v, ok := compatibilityKeys[key]; ok {
		return v
	}
	return key
}

var lightLiftPercent = map[api.ProgramWeek]float64{
	api.ProgramWeekWeek1: 0.62,
	api.ProgramWeekWeek2: 0.57,
	api.ProgramWeekWeek3: 0.66,
	api.ProgramWeekWeek4: 0.64,
	api.ProgramWeekWeek5: 0.59,
	api.ProgramWeekWeek6: 0.66,
	api.ProgramWeekWeek7: 0.7,
}

var lightBenchPercent = map[api.ProgramWeek]float64{
	api.ProgramWeekWeek1: 0.6,
	api.ProgramWeekWeek2: 0.65,
	api.ProgramWeekWeek3: 0.6,
	api.ProgramWeekWeek4: 0.65,
	api.ProgramWeekWeek5: 0.65,
	api.ProgramWeekWeek6: 0.65,
	api.ProgramWeekWeek7: 0.65,
}

var lightLiftPrescription = map[api.ProgramWeek]string{
	api.ProgramWeekWeek1: "5x5",
	api.ProgramWeekWeek2: "4x6",
	api.ProgramWeekWeek3: "5x4",
	api.ProgramWeekWeek4: "5x5",
	api.ProgramWeekWeek5: "4x6",
	api.ProgramWeekWeek6: "5x4",
	api.ProgramWeekWeek7: "3x3",
}

var armSetsByWeek = map[api.ProgramWeek]int{
	api.ProgramWeekWeek1: 2,
	api.ProgramWeekWeek2: 3,
	api.ProgramWeekWeek3: 3,
	api.ProgramWeekWeek4: 4,
	api.ProgramWeekWeek5: 2,
	api.ProgramWeekWeek6: 3,
	api.ProgramWeekWeek7: 3,
}

var deadliftAssistance = []ExerciseOption{
	{"good_morning", "Гуд-морнинг"},
	{"romanian_deadlift", "Румынская тяга"},
	{"deficit_deadlift", "Тяга из ямы"},
	{"deadlift", "Классическая становая тяга"},
	{"sumo_deadlift", "Становая тяга сумо"},
	{"paused_deadlift", "Становая тяга с паузами"},
}

var benchAssistance = []ExerciseOption{
	{"close_grip_bench", "Жим узким хватом"},
	{"reverse_grip_bench", "Жим обратным хватом"},
	{"incline_bench", "Жим на наклонной скамье"},
	{"dumbbell_bench", "Жим гантелей лежа"},
	{"dips", "Брусья"},
}

var squatAssistance = []ExerciseOption{
	{"zercher_squat", "Приседания Зерчера"},
	{"front_squat", "Приседания со штангой на груди"},
	{"high_bar_squat", "Приседания с высоким грифом"},
	{"low_bar_squat", "Приседания с низким грифом"},
	{"bulgarian_split_squat", "Болгарские сплит-приседания"},
}

var gppAbs = []ExerciseOption{{"abs", "Упражнение на пресс"}}
var gppTriceps = []ExerciseOption{{"triceps", "Трицепс"}}
var gppHorizontalPull = []ExerciseOption{
	{"barbell_row", "Тяга штанги в наклоне"},
	{"cable_seated_row", "Горизонтальный блок"},
	{"dumbbell_row", "Тяга гантели в наклоне"},
	{"lever_horizontal_row", "Рычажная горизонтальная тяга"},
}
var gppBiceps = []ExerciseOption{{"biceps", "Бицепс"}}
var gppVerticalPull = []ExerciseOption{
	{"pull_up", "Подтягивания"},
	{"lat_pulldown", "Вертикальный блок"},
	{"lever_vertical_row", "Рычажная вертикальная тяга"},
}
var gppOverheadPress = []ExerciseOption{
	{"dumbbell_military_press", "Армейский жим гантелей"},
	{"handstand_push_up", "Отжимания в стойке на руках"},
	{"kettlebell_military_press", "Армейский жим гирь"},
	{"one_arm_military_press", "Армейский жим одной рукой"},
	{"barbell_military_press", "Армейский жим штанги"},
}

var compatibilityKeys = map[string]string{
	"close_grip_bench_press":    "close_grip_bench",
	"reverse_grip_bench_press":  "reverse_grip_bench",
	"incline_bench_press":       "incline_bench",
	"dumbbell_bench_press":      "dumbbell_bench",
	"dip":                       "dips",
	"bent_over_barbell_row":     "barbell_row",
	"seated_cable_row":          "cable_seated_row",
	"one_arm_dumbbell_row":      "dumbbell_row",
	"dumbbell_overhead_press":   "dumbbell_military_press",
	"kettlebell_overhead_press": "kettlebell_military_press",
	"one_arm_overhead_press":    "one_arm_military_press",
	"barbell_overhead_press":    "barbell_military_press",
}

var optionLabels = func() map[string]string {
	sets := [][]ExerciseOption{
		deadliftAssistance, benchAssistance, squatAssistance,
		gppAbs, gppTriceps, gppHorizontalPull, gppBiceps, gppVerticalPull, gppOverheadPress,
		{{"deadlift", "Становая тяга"}, {"bench_press", "Жим лежа"}, {"squat", "Приседания"}},
	}
	out := make(map[string]string)
	for _, set := range sets {
		for _, opt := range set {
			out[opt.ID] = opt.Label
		}
	}
	return out
}()
