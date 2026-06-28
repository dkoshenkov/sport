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
			Abs:            toAPIOptions(gppAbs),
			Triceps:        toAPIOptions(gppTriceps),
			HorizontalPull: toAPIOptions(gppHorizontalPull),
			Biceps:         toAPIOptions(gppBiceps),
			VerticalPull:   toAPIOptions(gppVerticalPull),
			OverheadPress:  toAPIOptions(gppOverheadPress),
		},
	}
}

func Calculate(selection api.ProgramSelection) (*api.TrainingPlanResponse, error) {
	settings := selection.Settings
	if err := validateSettings(settings); err != nil {
		return nil, err
	}
	multipliers := map[api.ProgramWeek]int{
		api.ProgramWeekWeek1: -4,
		api.ProgramWeekWeek2: -3,
		api.ProgramWeekWeek3: -2,
		api.ProgramWeekWeek4: -1,
		api.ProgramWeekWeek5: 0,
		api.ProgramWeekWeek6: 1,
		api.ProgramWeekWeek7: 2,
	}

	mainPrescription := func(max float64) api.Prescription {
		if selection.Week == api.ProgramWeekWeek8 {
			return api.Prescription{
				SetsRepsText: "1ПМ",
				WeightText:   api.NewOptNilString("1ПМ"),
				Unit:         api.NewOptNilString("кг"),
			}
		}
		weight := roundToNearest2_5(baseWeight(max, settings.Variant) + progressionStep(max, settings.ProgressionStep)*float64(multipliers[selection.Week]))
		return api.Prescription{
			SetsRepsText: "5x5",
			WeightKg:     api.NewOptNilFloat64(weight),
			Unit:         api.NewOptNilString("кг"),
		}
	}

	days := []api.TrainingDay{
		{
			ID:    api.TrainingDayIDDay1,
			Label: "День 1",
			Rows: compactRows([]api.TrainingRow{
				row("day_1_deadlift_main", "deadlift", "Становая тяга", api.TrainingRowKindMain, mainPrescription(settings.OneRepMaxKg.Deadlift)),
				row("day_1_deadlift_assistance", settings.Assistance.Deadlift, label(settings.Assistance.Deadlift), api.TrainingRowKindAssistance, assistancePrescription()),
				gppRow("day_1_horizontal_pull", settings.Gpp.HorizontalPull, "3x8-12"),
				gppRow("day_1_abs", settings.Gpp.Abs, "3x10-20"),
			}),
		},
		{
			ID:    api.TrainingDayIDDay2,
			Label: "День 2",
			Rows: compactRows([]api.TrainingRow{
				row("day_2_bench_main", "bench_press", "Жим лежа", api.TrainingRowKindMain, mainPrescription(settings.OneRepMaxKg.Bench)),
				row("day_2_bench_assistance", settings.Assistance.Bench, label(settings.Assistance.Bench), api.TrainingRowKindAssistance, assistancePrescription()),
				gppRow("day_2_vertical_pull", settings.Gpp.VerticalPull, "3x8-12"),
				gppRow("day_2_triceps", settings.Gpp.Triceps, "3x10-15"),
			}),
		},
		{
			ID:    api.TrainingDayIDDay3,
			Label: "День 3",
			Rows: compactRows([]api.TrainingRow{
				row("day_3_squat_main", "squat", "Приседания", api.TrainingRowKindMain, mainPrescription(settings.OneRepMaxKg.Squat)),
				row("day_3_squat_assistance", settings.Assistance.Squat, label(settings.Assistance.Squat), api.TrainingRowKindAssistance, assistancePrescription()),
				gppRow("day_3_overhead_press", settings.Gpp.OverheadPress, "3x8-12"),
				gppRow("day_3_biceps", settings.Gpp.Biceps, "3x10-15"),
			}),
		},
	}

	warnings := []string{}
	if settings.ProgressionStep == api.ProgressionStepStep5Percent {
		warnings = append(warnings, "Используется явный шаг прогрессии 5%; XLSX-совместимое значение по умолчанию - 4%.")
	}

	return &api.TrainingPlanResponse{
		Selection:      selection,
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
			Deadlift: "romanian_deadlift",
			Bench:    "close_grip_bench_press",
			Squat:    "front_squat",
		},
		Gpp: api.GPPSelection{
			Abs:            api.NewNilString("abs"),
			Triceps:        api.NewNilString("triceps"),
			HorizontalPull: api.NewNilString("bent_over_barbell_row"),
			Biceps:         api.NewNilString("biceps"),
			VerticalPull:   api.NewNilString("pull_up"),
			OverheadPress:  api.NewNilString("dumbbell_overhead_press"),
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
		if key, ok := opt.Get(); ok {
			if _, exists := optionLabels[key]; !exists {
				return fmt.Errorf("unknown gpp option %q", key)
			}
		}
	}
	return nil
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

func gppRow(rowID string, opt api.NilString, setsReps string) api.TrainingRow {
	key, ok := opt.Get()
	if !ok || key == "" {
		return api.TrainingRow{}
	}
	return row(rowID, key, label(key), api.TrainingRowKindGpp, api.Prescription{
		SetsRepsText: setsReps,
		RpeText:      api.NewOptNilString("RPE 6-7"),
	})
}

func assistancePrescription() api.Prescription {
	return api.Prescription{
		SetsRepsText: "4x6-10",
		RpeText:      api.NewOptNilString("RPE 7"),
	}
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

func label(key string) string {
	if v, ok := optionLabels[key]; ok {
		return v
	}
	return key
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
	{"close_grip_bench_press", "Жим узким хватом"},
	{"reverse_grip_bench_press", "Жим обратным хватом"},
	{"incline_bench_press", "Жим на наклонной скамье"},
	{"dumbbell_bench_press", "Жим гантелей лежа"},
	{"dip", "Брусья"},
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
	{"bent_over_barbell_row", "Тяга штанги в наклоне"},
	{"seated_cable_row", "Горизонтальный блок"},
	{"one_arm_dumbbell_row", "Тяга гантели в наклоне"},
	{"lever_horizontal_row", "Рычажная горизонтальная тяга"},
}
var gppBiceps = []ExerciseOption{{"biceps", "Бицепс"}}
var gppVerticalPull = []ExerciseOption{
	{"pull_up", "Подтягивания"},
	{"lat_pulldown", "Вертикальный блок"},
	{"lever_vertical_row", "Рычажная вертикальная тяга"},
}
var gppOverheadPress = []ExerciseOption{
	{"dumbbell_overhead_press", "Армейский жим гантелей"},
	{"handstand_push_up", "Отжимания в стойке на руках"},
	{"kettlebell_overhead_press", "Армейский жим гирь"},
	{"one_arm_overhead_press", "Армейский жим одной рукой"},
	{"barbell_overhead_press", "Армейский жим штанги"},
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
