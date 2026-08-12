package app

import (
	"testing"

	"sport/server/internal/api"
)

func TestProgressDayDoneRequiresEveryExerciseToBeDone(t *testing.T) {
	plan := &api.TrainingPlanResponse{Days: []api.TrainingDay{{
		ID: api.TrainingDayIDDay1,
		Rows: []api.TrainingRow{
			{ExerciseKey: "deadlift", Kind: api.TrainingRowKindMain},
			{ExerciseKey: "bench_press", Kind: api.TrainingRowKindLight},
		},
	}}}

	tests := []struct {
		name        string
		checkpoints []api.ProgressCheckpoint
		want        bool
	}{
		{name: "missing checkpoint", checkpoints: []api.ProgressCheckpoint{{DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "deadlift", Status: api.CheckpointStatusDone}}, want: false},
		{name: "partial does not close day", checkpoints: []api.ProgressCheckpoint{{DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "deadlift", Status: api.CheckpointStatusDone}, {DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "bench_press", Status: api.CheckpointStatusPartial}}, want: false},
		{name: "skipped does not close day", checkpoints: []api.ProgressCheckpoint{{DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "deadlift", Status: api.CheckpointStatusDone}, {DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "bench_press", Status: api.CheckpointStatusSkipped}}, want: false},
		{name: "all exercises done", checkpoints: []api.ProgressCheckpoint{{DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "deadlift", Status: api.CheckpointStatusDone}, {DayId: api.ProgressCheckpointDayIdDay1, ExerciseKey: "bench_press", Status: api.CheckpointStatusDone}}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := progressDayDone(plan.Days[0], test.checkpoints); got != test.want {
				t.Fatalf("progressDayDone() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNextProgramWeek(t *testing.T) {
	for _, test := range []struct {
		week api.ProgramWeek
		want api.ProgramWeek
		ok   bool
	}{
		{week: api.ProgramWeekWeek1, want: api.ProgramWeekWeek2, ok: true},
		{week: api.ProgramWeekWeek7, want: api.ProgramWeekWeek8, ok: true},
		{week: api.ProgramWeekWeek8, ok: false},
	} {
		t.Run(string(test.week), func(t *testing.T) {
			got, ok := nextProgramWeek(test.week)
			if got != test.want || ok != test.ok {
				t.Fatalf("nextProgramWeek() = %q, %v; want %q, %v", got, ok, test.want, test.ok)
			}
		})
	}
}
