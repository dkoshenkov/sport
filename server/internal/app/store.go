package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sport/server/internal/api"
	"sport/server/internal/program"
)

type Session struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt time.Time
}

type Store interface {
	CreateUser(ctx context.Context, nickname, passwordHash string) (api.User, error)
	UserByNickname(ctx context.Context, nickname string) (api.User, string, bool, error)
	UserByID(ctx context.Context, id uuid.UUID) (api.User, bool, error)
	CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	Session(ctx context.Context, tokenHash string) (Session, bool, error)
	RevokeSession(ctx context.Context, tokenHash string) error
	Profile(ctx context.Context, userID uuid.UUID) (api.AthleteProfile, error)
	SaveProfile(ctx context.Context, userID uuid.UUID, input api.AthleteProfileInput) (api.AthleteProfile, error)
	ListCycles(ctx context.Context, userID uuid.UUID) ([]api.ProgramCycle, uuid.UUID, bool, error)
	CreateCycle(ctx context.Context, userID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, error)
	CurrentCycle(ctx context.Context, userID uuid.UUID) (api.ProgramCycle, bool, error)
	SaveCurrentCycle(ctx context.Context, userID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, error)
	SaveCycle(ctx context.Context, userID, cycleID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, bool, error)
	ActivateCycle(ctx context.Context, userID, cycleID uuid.UUID) (api.ProgramCycle, bool, error)
	AdvanceCycle(ctx context.Context, userID uuid.UUID, week api.ProgramWeek) (api.ProgramCycle, bool, error)
	ListProgress(ctx context.Context, cycleID uuid.UUID, week api.ProgramWeek) ([]api.ProgressCheckpoint, error)
	UpsertProgress(ctx context.Context, cycleID uuid.UUID, input api.ProgressCheckpointInput) (api.ProgressCheckpoint, error)
	DeleteProgress(ctx context.Context, cycleID, checkpointID uuid.UUID) (bool, error)
	ExerciseDetails(ctx context.Context, exerciseKey string) (api.ExerciseDetails, bool, error)
	ListExercises(ctx context.Context, params api.ListExercisesParams) (api.ExerciseCatalogListResponse, error)
	CatalogExercise(ctx context.Context, datasetExerciseID string) (api.ExerciseCatalogItem, bool, error)
}

func defaultProfile(now time.Time) api.AthleteProfile {
	settings := program.DefaultSettings()
	return api.AthleteProfile{
		OneRepMaxKg: api.ProfileOneRepMaxes{
			Deadlift: api.NewNilFloat64(settings.OneRepMaxKg.Deadlift),
			Bench:    api.NewNilFloat64(settings.OneRepMaxKg.Bench),
			Squat:    api.NewNilFloat64(settings.OneRepMaxKg.Squat),
		},
		PreferredVariant:         settings.Variant,
		PreferredProgressionStep: settings.ProgressionStep,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
}
