package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"sport/server/internal/api"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateUser(ctx context.Context, nickname, passwordHash string) (api.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return api.User{}, fmt.Errorf("begin create user: %w", err)
	}
	defer tx.Rollback(ctx)

	var user api.User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (nickname, password_hash)
		VALUES ($1, $2)
		RETURNING id, nickname, created_at
	`, nickname, passwordHash).Scan(&user.ID, &user.Nickname, &user.CreatedAt)
	if isUniqueViolation(err) {
		return api.User{}, errConflict
	}
	if err != nil {
		return api.User{}, fmt.Errorf("insert user: %w", err)
	}

	profile := defaultProfile(time.Now().UTC())
	if _, err := tx.Exec(ctx, `
		INSERT INTO athlete_profiles (
			user_id, deadlift_1rm_kg, bench_1rm_kg, squat_1rm_kg,
			preferred_variant, preferred_progression_step, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID,
		nilFloatValue(profile.OneRepMaxKg.Deadlift),
		nilFloatValue(profile.OneRepMaxKg.Bench),
		nilFloatValue(profile.OneRepMaxKg.Squat),
		profile.PreferredVariant,
		profile.PreferredProgressionStep,
		optNilStringValue(profile.Notes),
	); err != nil {
		return api.User{}, fmt.Errorf("insert default profile: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return api.User{}, fmt.Errorf("commit create user: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) UserByNickname(ctx context.Context, nickname string) (api.User, string, bool, error) {
	var user api.User
	var passwordHash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, nickname, created_at, password_hash
		FROM users
		WHERE nickname = $1
	`, nickname).Scan(&user.ID, &user.Nickname, &user.CreatedAt, &passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return api.User{}, "", false, nil
	}
	if err != nil {
		return api.User{}, "", false, fmt.Errorf("select user by nickname: %w", err)
	}
	return user, passwordHash, true, nil
}

func (s *PostgresStore) UserByID(ctx context.Context, id uuid.UUID) (api.User, bool, error) {
	var user api.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, nickname, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Nickname, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return api.User{}, false, nil
	}
	if err != nil {
		return api.User{}, false, fmt.Errorf("select user by id: %w", err)
	}
	return user, true, nil
}

func (s *PostgresStore) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO auth_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt.UTC())
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *PostgresStore) Session(ctx context.Context, tokenHash string) (Session, bool, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `
		UPDATE auth_sessions
		SET last_seen_at = now()
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > now()
		RETURNING user_id, expires_at
	`, tokenHash).Scan(&session.UserID, &session.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, fmt.Errorf("select session: %w", err)
	}
	return session, true, nil
}

func (s *PostgresStore) RevokeSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (s *PostgresStore) Profile(ctx context.Context, userID uuid.UUID) (api.AthleteProfile, error) {
	profile, err := scanProfile(s.pool.QueryRow(ctx, profileQuery()+` WHERE user_id = $1`, userID))
	if err != nil {
		return api.AthleteProfile{}, fmt.Errorf("select profile: %w", err)
	}
	return profile, nil
}

func (s *PostgresStore) SaveProfile(ctx context.Context, userID uuid.UUID, input api.AthleteProfileInput) (api.AthleteProfile, error) {
	profile, err := scanProfile(s.pool.QueryRow(ctx, `
		INSERT INTO athlete_profiles (
			user_id, deadlift_1rm_kg, bench_1rm_kg, squat_1rm_kg,
			preferred_variant, preferred_progression_step, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			deadlift_1rm_kg = EXCLUDED.deadlift_1rm_kg,
			bench_1rm_kg = EXCLUDED.bench_1rm_kg,
			squat_1rm_kg = EXCLUDED.squat_1rm_kg,
			preferred_variant = EXCLUDED.preferred_variant,
			preferred_progression_step = EXCLUDED.preferred_progression_step,
			notes = EXCLUDED.notes,
			updated_at = now()
		RETURNING deadlift_1rm_kg::float8, bench_1rm_kg::float8, squat_1rm_kg::float8,
			preferred_variant, preferred_progression_step, notes, created_at, updated_at
	`, userID,
		nilFloatValue(input.OneRepMaxKg.Deadlift),
		nilFloatValue(input.OneRepMaxKg.Bench),
		nilFloatValue(input.OneRepMaxKg.Squat),
		input.PreferredVariant,
		input.PreferredProgressionStep,
		optNilStringValue(input.Notes),
	))
	if err != nil {
		return api.AthleteProfile{}, fmt.Errorf("save profile: %w", err)
	}
	return profile, nil
}

func (s *PostgresStore) CurrentCycle(ctx context.Context, userID uuid.UUID) (api.ProgramCycle, bool, error) {
	cycle, err := scanCycle(s.pool.QueryRow(ctx, cycleQuery()+` WHERE c.user_id = $1 AND c.status = 'active'`+cycleGroupBy(), userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return api.ProgramCycle{}, false, nil
	}
	if err != nil {
		return api.ProgramCycle{}, false, fmt.Errorf("select current cycle: %w", err)
	}
	return cycle, true, nil
}

func (s *PostgresStore) SaveCurrentCycle(ctx context.Context, userID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return api.ProgramCycle{}, fmt.Errorf("begin save cycle: %w", err)
	}
	defer tx.Rollback(ctx)

	var cycleID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO program_cycles (user_id, title, status, current_week, started_at)
		VALUES ($1, $2, 'active', $3, now())
		ON CONFLICT (user_id) WHERE status = 'active' DO UPDATE SET
			title = EXCLUDED.title,
			current_week = EXCLUDED.current_week,
			updated_at = now()
		RETURNING id
	`, userID, title, week).Scan(&cycleID)
	if err != nil {
		return api.ProgramCycle{}, fmt.Errorf("upsert cycle: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO cycle_settings (
			cycle_id, deadlift_1rm_kg, bench_1rm_kg, squat_1rm_kg,
			variant, progression_step, deadlift_assistance, bench_assistance, squat_assistance,
			gpp_abs, gpp_triceps, gpp_horizontal_pull, gpp_biceps, gpp_vertical_pull, gpp_overhead_press
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (cycle_id) DO UPDATE SET
			deadlift_1rm_kg = EXCLUDED.deadlift_1rm_kg,
			bench_1rm_kg = EXCLUDED.bench_1rm_kg,
			squat_1rm_kg = EXCLUDED.squat_1rm_kg,
			variant = EXCLUDED.variant,
			progression_step = EXCLUDED.progression_step,
			deadlift_assistance = EXCLUDED.deadlift_assistance,
			bench_assistance = EXCLUDED.bench_assistance,
			squat_assistance = EXCLUDED.squat_assistance,
			gpp_abs = EXCLUDED.gpp_abs,
			gpp_triceps = EXCLUDED.gpp_triceps,
			gpp_horizontal_pull = EXCLUDED.gpp_horizontal_pull,
			gpp_biceps = EXCLUDED.gpp_biceps,
			gpp_vertical_pull = EXCLUDED.gpp_vertical_pull,
			gpp_overhead_press = EXCLUDED.gpp_overhead_press,
			updated_at = now()
	`, cycleID,
		settings.OneRepMaxKg.Deadlift,
		settings.OneRepMaxKg.Bench,
		settings.OneRepMaxKg.Squat,
		settings.Variant,
		settings.ProgressionStep,
		settings.Assistance.Deadlift,
		settings.Assistance.Bench,
		settings.Assistance.Squat,
		nilStringValue(settings.Gpp.Abs),
		nilStringValue(settings.Gpp.Triceps),
		nilStringValue(settings.Gpp.HorizontalPull),
		nilStringValue(settings.Gpp.Biceps),
		nilStringValue(settings.Gpp.VerticalPull),
		nilStringValue(settings.Gpp.OverheadPress),
	); err != nil {
		return api.ProgramCycle{}, fmt.Errorf("upsert cycle settings: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return api.ProgramCycle{}, fmt.Errorf("commit save cycle: %w", err)
	}
	cycle, ok, err := s.CurrentCycle(ctx, userID)
	if err != nil {
		return api.ProgramCycle{}, err
	}
	if !ok {
		return api.ProgramCycle{}, errNotFound
	}
	return cycle, nil
}

func (s *PostgresStore) AdvanceCycle(ctx context.Context, userID uuid.UUID, week api.ProgramWeek) (api.ProgramCycle, bool, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE program_cycles
		SET current_week = $2, updated_at = now()
		WHERE user_id = $1 AND status = 'active'
	`, userID, week)
	if err != nil {
		return api.ProgramCycle{}, false, fmt.Errorf("advance cycle: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return api.ProgramCycle{}, false, nil
	}
	return s.CurrentCycle(ctx, userID)
}

func (s *PostgresStore) ListProgress(ctx context.Context, cycleID uuid.UUID, week api.ProgramWeek) ([]api.ProgressCheckpoint, error) {
	args := []any{cycleID}
	query := checkpointQuery() + ` WHERE cycle_id = $1`
	if week != "" {
		args = append(args, week)
		query += ` AND week = $2`
	}
	query += ` ORDER BY week, day_id, exercise_key`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	defer rows.Close()

	var items []api.ProgressCheckpoint
	for rows.Next() {
		checkpoint, err := scanCheckpoint(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, checkpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate progress: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) UpsertProgress(ctx context.Context, cycleID uuid.UUID, input api.ProgressCheckpointInput) (api.ProgressCheckpoint, error) {
	prescribed := input.Prescribed.Or(api.CheckpointPrescriptionSnapshot{})
	completed := input.Completed.Or(api.CheckpointCompletedData{})
	checkpoint, err := scanCheckpoint(s.pool.QueryRow(ctx, `
		INSERT INTO progress_checkpoints (
			cycle_id, week, day_id, exercise_key, row_kind, status,
			prescribed_sets, prescribed_reps, prescribed_weight_kg, prescribed_rpe,
			completed_sets, completed_reps, actual_weight_kg, actual_rpe,
			notes, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (cycle_id, week, day_id, exercise_key) DO UPDATE SET
			row_kind = EXCLUDED.row_kind,
			status = EXCLUDED.status,
			prescribed_sets = EXCLUDED.prescribed_sets,
			prescribed_reps = EXCLUDED.prescribed_reps,
			prescribed_weight_kg = EXCLUDED.prescribed_weight_kg,
			prescribed_rpe = EXCLUDED.prescribed_rpe,
			completed_sets = EXCLUDED.completed_sets,
			completed_reps = EXCLUDED.completed_reps,
			actual_weight_kg = EXCLUDED.actual_weight_kg,
			actual_rpe = EXCLUDED.actual_rpe,
			notes = EXCLUDED.notes,
			completed_at = EXCLUDED.completed_at,
			updated_at = now()
		RETURNING id, week, day_id, exercise_key, row_kind, status,
			prescribed_sets, prescribed_reps, prescribed_weight_kg::float8, prescribed_rpe,
			completed_sets, completed_reps, actual_weight_kg::float8, actual_rpe,
			notes, completed_at, created_at, updated_at
	`, cycleID,
		input.Week,
		input.DayId,
		input.ExerciseKey,
		input.RowKind,
		input.Status,
		optNilIntValue(prescribed.Sets),
		optNilStringValue(prescribed.RepsText),
		optNilFloatValue(prescribed.WeightKg),
		optNilStringValue(prescribed.RpeText),
		optNilIntValue(completed.Sets),
		optNilStringValue(completed.RepsText),
		optNilFloatValue(completed.WeightKg),
		optNilStringValue(completed.RpeText),
		optNilStringValue(input.Notes),
		optNilTimeValue(input.CompletedAt),
	))
	if err != nil {
		return api.ProgressCheckpoint{}, fmt.Errorf("upsert progress: %w", err)
	}
	return checkpoint, nil
}

func (s *PostgresStore) DeleteProgress(ctx context.Context, cycleID, checkpointID uuid.UUID) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM progress_checkpoints
		WHERE cycle_id = $1 AND id = $2
	`, cycleID, checkpointID)
	if err != nil {
		return false, fmt.Errorf("delete progress: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func profileQuery() string {
	return `
		SELECT deadlift_1rm_kg::float8, bench_1rm_kg::float8, squat_1rm_kg::float8,
			preferred_variant, preferred_progression_step, notes, created_at, updated_at
		FROM athlete_profiles
	`
}

func scanProfile(row pgx.Row) (api.AthleteProfile, error) {
	var deadlift, bench, squat pgtype.Float8
	var notes pgtype.Text
	var profile api.AthleteProfile
	if err := row.Scan(
		&deadlift,
		&bench,
		&squat,
		&profile.PreferredVariant,
		&profile.PreferredProgressionStep,
		&notes,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return api.AthleteProfile{}, err
	}
	profile.OneRepMaxKg = api.ProfileOneRepMaxes{
		Deadlift: floatToNil(deadlift),
		Bench:    floatToNil(bench),
		Squat:    floatToNil(squat),
	}
	profile.Notes = textToOptNil(notes)
	return profile, nil
}

func cycleQuery() string {
	return `
		SELECT c.id, c.title, c.status, c.current_week,
			cs.deadlift_1rm_kg::float8, cs.bench_1rm_kg::float8, cs.squat_1rm_kg::float8,
			cs.variant, cs.progression_step,
			cs.deadlift_assistance, cs.bench_assistance, cs.squat_assistance,
			cs.gpp_abs, cs.gpp_triceps, cs.gpp_horizontal_pull, cs.gpp_biceps, cs.gpp_vertical_pull, cs.gpp_overhead_press,
			COALESCE(SUM(CASE WHEN p.status = 'done' THEN 1 ELSE 0 END), 0)::int,
			COALESCE(SUM(CASE WHEN p.status = 'partial' THEN 1 ELSE 0 END), 0)::int,
			COALESCE(SUM(CASE WHEN p.status = 'skipped' THEN 1 ELSE 0 END), 0)::int,
			COALESCE(SUM(CASE WHEN p.status = 'planned' THEN 1 ELSE 0 END), 0)::int,
			c.started_at, c.completed_at, c.created_at, c.updated_at
		FROM program_cycles c
		JOIN cycle_settings cs ON cs.cycle_id = c.id
		LEFT JOIN progress_checkpoints p ON p.cycle_id = c.id
	`
}

func cycleGroupBy() string {
	return `
		GROUP BY c.id, cs.cycle_id
	`
}

func scanCycle(row pgx.Row) (api.ProgramCycle, error) {
	var cycle api.ProgramCycle
	var gppAbs, gppTriceps, gppHorizontalPull, gppBiceps, gppVerticalPull, gppOverheadPress pgtype.Text
	var startedAt, completedAt pgtype.Timestamptz
	if err := row.Scan(
		&cycle.ID,
		&cycle.Title,
		&cycle.Status,
		&cycle.CurrentWeek,
		&cycle.Settings.OneRepMaxKg.Deadlift,
		&cycle.Settings.OneRepMaxKg.Bench,
		&cycle.Settings.OneRepMaxKg.Squat,
		&cycle.Settings.Variant,
		&cycle.Settings.ProgressionStep,
		&cycle.Settings.Assistance.Deadlift,
		&cycle.Settings.Assistance.Bench,
		&cycle.Settings.Assistance.Squat,
		&gppAbs,
		&gppTriceps,
		&gppHorizontalPull,
		&gppBiceps,
		&gppVerticalPull,
		&gppOverheadPress,
		&cycle.ProgressSummary.Done,
		&cycle.ProgressSummary.Partial,
		&cycle.ProgressSummary.Skipped,
		&cycle.ProgressSummary.Planned,
		&startedAt,
		&completedAt,
		&cycle.CreatedAt,
		&cycle.UpdatedAt,
	); err != nil {
		return api.ProgramCycle{}, err
	}
	cycle.Settings.Gpp = api.GPPSelection{
		Abs:            textToNil(gppAbs),
		Triceps:        textToNil(gppTriceps),
		HorizontalPull: textToNil(gppHorizontalPull),
		Biceps:         textToNil(gppBiceps),
		VerticalPull:   textToNil(gppVerticalPull),
		OverheadPress:  textToNil(gppOverheadPress),
	}
	cycle.StartedAt = timeToOptNil(startedAt)
	cycle.CompletedAt = timeToOptNil(completedAt)
	return cycle, nil
}

func checkpointQuery() string {
	return `
		SELECT id, week, day_id, exercise_key, row_kind, status,
			prescribed_sets, prescribed_reps, prescribed_weight_kg::float8, prescribed_rpe,
			completed_sets, completed_reps, actual_weight_kg::float8, actual_rpe,
			notes, completed_at, created_at, updated_at
		FROM progress_checkpoints
	`
}

func scanCheckpoint(row pgx.Row) (api.ProgressCheckpoint, error) {
	var checkpoint api.ProgressCheckpoint
	var prescribedSets, completedSets pgtype.Int4
	var prescribedReps, prescribedRPE, completedReps, actualRPE, notes pgtype.Text
	var prescribedWeight, actualWeight pgtype.Float8
	var completedAt pgtype.Timestamptz
	if err := row.Scan(
		&checkpoint.ID,
		&checkpoint.Week,
		&checkpoint.DayId,
		&checkpoint.ExerciseKey,
		&checkpoint.RowKind,
		&checkpoint.Status,
		&prescribedSets,
		&prescribedReps,
		&prescribedWeight,
		&prescribedRPE,
		&completedSets,
		&completedReps,
		&actualWeight,
		&actualRPE,
		&notes,
		&completedAt,
		&checkpoint.CreatedAt,
		&checkpoint.UpdatedAt,
	); err != nil {
		return api.ProgressCheckpoint{}, err
	}
	checkpoint.Prescribed = api.NewOptCheckpointPrescriptionSnapshot(api.CheckpointPrescriptionSnapshot{
		Sets:     intToOptNil(prescribedSets),
		RepsText: textToOptNil(prescribedReps),
		WeightKg: floatToOptNil(prescribedWeight),
		RpeText:  textToOptNil(prescribedRPE),
	})
	checkpoint.Completed = api.NewOptCheckpointCompletedData(api.CheckpointCompletedData{
		Sets:     intToOptNil(completedSets),
		RepsText: textToOptNil(completedReps),
		WeightKg: floatToOptNil(actualWeight),
		RpeText:  textToOptNil(actualRPE),
	})
	checkpoint.Notes = textToOptNil(notes)
	checkpoint.CompletedAt = timeToOptNil(completedAt)
	return checkpoint, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nilFloatValue(value api.NilFloat64) any {
	if v, ok := value.Get(); ok {
		return v
	}
	return nil
}

func nilStringValue(value api.NilString) any {
	if v, ok := value.Get(); ok && v != "" {
		return v
	}
	return nil
}

func optNilStringValue(value api.OptNilString) any {
	if v, ok := value.Get(); ok {
		return v
	}
	return nil
}

func optNilFloatValue(value api.OptNilFloat64) any {
	if v, ok := value.Get(); ok {
		return v
	}
	return nil
}

func optNilIntValue(value api.OptNilInt) any {
	if v, ok := value.Get(); ok {
		return v
	}
	return nil
}

func optNilTimeValue(value api.OptNilDateTime) any {
	if v, ok := value.Get(); ok {
		return v
	}
	return nil
}

func floatToNil(value pgtype.Float8) api.NilFloat64 {
	if value.Valid {
		return api.NewNilFloat64(value.Float64)
	}
	var out api.NilFloat64
	out.SetToNull()
	return out
}

func textToNil(value pgtype.Text) api.NilString {
	if value.Valid {
		return api.NewNilString(value.String)
	}
	var out api.NilString
	out.SetToNull()
	return out
}

func textToOptNil(value pgtype.Text) api.OptNilString {
	if value.Valid {
		return api.NewOptNilString(value.String)
	}
	var out api.OptNilString
	out.SetToNull()
	return out
}

func floatToOptNil(value pgtype.Float8) api.OptNilFloat64 {
	if value.Valid {
		return api.NewOptNilFloat64(value.Float64)
	}
	var out api.OptNilFloat64
	out.SetToNull()
	return out
}

func intToOptNil(value pgtype.Int4) api.OptNilInt {
	if value.Valid {
		return api.NewOptNilInt(int(value.Int32))
	}
	var out api.OptNilInt
	out.SetToNull()
	return out
}

func timeToOptNil(value pgtype.Timestamptz) api.OptNilDateTime {
	if value.Valid {
		return api.NewOptNilDateTime(value.Time)
	}
	var out api.OptNilDateTime
	out.SetToNull()
	return out
}
