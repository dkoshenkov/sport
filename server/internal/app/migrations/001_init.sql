CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS athlete_profiles (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    deadlift_1rm_kg numeric NULL,
    bench_1rm_kg numeric NULL,
    squat_1rm_kg numeric NULL,
    preferred_variant text NOT NULL,
    preferred_progression_step text NOT NULL,
    notes text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NULL,
    revoked_at timestamptz NULL
);

CREATE INDEX IF NOT EXISTS auth_sessions_user_id_idx ON auth_sessions(user_id);

CREATE TABLE IF NOT EXISTS program_cycles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title text NOT NULL,
    status text NOT NULL,
    current_week text NOT NULL,
    started_at timestamptz NULL,
    completed_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS program_cycles_active_user_idx ON program_cycles(user_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS cycle_settings (
    cycle_id uuid PRIMARY KEY REFERENCES program_cycles(id) ON DELETE CASCADE,
    deadlift_1rm_kg numeric NOT NULL,
    bench_1rm_kg numeric NOT NULL,
    squat_1rm_kg numeric NOT NULL,
    variant text NOT NULL,
    progression_step text NOT NULL,
    deadlift_assistance text NOT NULL,
    bench_assistance text NOT NULL,
    squat_assistance text NOT NULL,
    gpp_abs text NULL,
    gpp_triceps text NULL,
    gpp_horizontal_pull text NULL,
    gpp_biceps text NULL,
    gpp_vertical_pull text NULL,
    gpp_overhead_press text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS progress_checkpoints (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id uuid NOT NULL REFERENCES program_cycles(id) ON DELETE CASCADE,
    week text NOT NULL,
    day_id text NOT NULL,
    exercise_key text NOT NULL,
    row_kind text NOT NULL,
    status text NOT NULL,
    prescribed_sets integer NULL,
    prescribed_reps text NULL,
    prescribed_weight_kg numeric NULL,
    prescribed_rpe text NULL,
    completed_sets integer NULL,
    completed_reps text NULL,
    actual_weight_kg numeric NULL,
    actual_rpe text NULL,
    notes text NULL,
    completed_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (cycle_id, week, day_id, exercise_key)
);
