package app

import "sport/server/internal/api"

// ProgressRequirement identifies one exercise that must be completed for a
// week to be considered complete.
type ProgressRequirement struct {
	DayID       string
	ExerciseKey string
}

func progressRequirements(plan *api.TrainingPlanResponse) []ProgressRequirement {
	var requirements []ProgressRequirement
	for _, day := range plan.Days {
		for _, row := range day.Rows {
			requirements = append(requirements, ProgressRequirement{
				DayID:       string(day.ID),
				ExerciseKey: row.ExerciseKey,
			})
		}
	}
	return requirements
}

func progressDayDone(day api.TrainingDay, checkpoints []api.ProgressCheckpoint) bool {
	if len(day.Rows) == 0 {
		return false
	}

	done := make(map[string]bool, len(checkpoints))
	for _, checkpoint := range checkpoints {
		if checkpoint.Status == api.CheckpointStatusDone {
			done[checkpointKey(string(checkpoint.DayId), checkpoint.ExerciseKey)] = true
		}
	}
	for _, row := range day.Rows {
		if !done[checkpointKey(string(day.ID), row.ExerciseKey)] {
			return false
		}
	}
	return true
}

func progressWeekDone(plan *api.TrainingPlanResponse, checkpoints []api.ProgressCheckpoint) bool {
	if len(plan.Days) == 0 {
		return false
	}
	for _, day := range plan.Days {
		if !progressDayDone(day, checkpoints) {
			return false
		}
	}
	return true
}

func checkpointKey(dayID, exerciseKey string) string {
	return dayID + "\x00" + exerciseKey
}

func nextProgramWeek(week api.ProgramWeek) (api.ProgramWeek, bool) {
	switch week {
	case api.ProgramWeekWeek1:
		return api.ProgramWeekWeek2, true
	case api.ProgramWeekWeek2:
		return api.ProgramWeekWeek3, true
	case api.ProgramWeekWeek3:
		return api.ProgramWeekWeek4, true
	case api.ProgramWeekWeek4:
		return api.ProgramWeekWeek5, true
	case api.ProgramWeekWeek5:
		return api.ProgramWeekWeek6, true
	case api.ProgramWeekWeek6:
		return api.ProgramWeekWeek7, true
	case api.ProgramWeekWeek7:
		return api.ProgramWeekWeek8, true
	default:
		return "", false
	}
}

func checkpointBelongsToPlan(plan *api.TrainingPlanResponse, input api.ProgressCheckpointInput) bool {
	for _, day := range plan.Days {
		if string(day.ID) != string(input.DayId) {
			continue
		}
		for _, row := range day.Rows {
			if row.ExerciseKey == input.ExerciseKey && row.Kind == input.RowKind {
				return true
			}
		}
	}
	return false
}
