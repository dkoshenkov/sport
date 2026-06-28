# Data Model

The service is stateful and uses a database.

Auth is local to this service for the MVP:

- user identity is a unique `nickname`;
- password is accepted only during register/login;
- store only a password hash;
- authenticated requests use an HttpOnly `sid` cookie.

The JavaScript-readable `init=1` cookie remains a bootstrap marker only. It is not authentication.

## Tables

### users

| Column          | Type        | Notes                   |
|-----------------|-------------|-------------------------|
| `id`            | uuid        | Internal DB id          |
| `nickname`      | text unique | Public login name       |
| `password_hash` | text        | Argon2id or bcrypt hash |
| `created_at`    | timestamptz |                         |
| `updated_at`    | timestamptz |                         |

### athlete_profiles

Default athlete data and program preferences.

Profile values are used as defaults when creating/updating an active cycle. Active cycle settings store their own snapshot, so changing the profile does not silently rewrite an in-progress or completed cycle.

| Column                       | Type             | Notes                                |
|------------------------------|------------------|--------------------------------------|
| `user_id`                    | uuid primary key | FK `users.id`                        |
| `deadlift_1rm_kg`            | numeric nullable | Default deadlift 1RM                 |
| `bench_1rm_kg`               | numeric nullable | Default bench press 1RM              |
| `squat_1rm_kg`               | numeric nullable | Default squat 1RM                    |
| `preferred_variant`          | text             | `variant_1` or `variant_2`           |
| `preferred_progression_step` | text             | `step_4_percent` or `step_5_percent` |
| `notes`                      | text nullable    | Optional user notes                  |
| `created_at`                 | timestamptz      |                                      |
| `updated_at`                 | timestamptz      |                                      |

### auth_sessions

| Column         | Type                 | Notes                             |
|----------------|----------------------|-----------------------------------|
| `id`           | uuid                 | Internal session id               |
| `user_id`      | uuid                 | FK `users.id`                     |
| `token_hash`   | text unique          | Hash of opaque `sid` cookie value |
| `expires_at`   | timestamptz          |                                   |
| `created_at`   | timestamptz          |                                   |
| `last_seen_at` | timestamptz nullable |                                   |
| `revoked_at`   | timestamptz nullable | Set on logout                     |

### program_cycles

One user can have several cycles over time. The MVP can expose only the active cycle.

| Column         | Type                 | Notes                             |
|----------------|----------------------|-----------------------------------|
| `id`           | uuid                 |                                   |
| `user_id`      | uuid                 | FK `users.id`                     |
| `title`        | text                 | User-facing cycle name            |
| `status`       | text                 | `active`, `completed`, `archived` |
| `current_week` | text                 | `week_1` ... `week_8`             |
| `started_at`   | timestamptz nullable |                                   |
| `completed_at` | timestamptz nullable |                                   |
| `created_at`   | timestamptz          |                                   |
| `updated_at`   | timestamptz          |                                   |

Recommended index:

- unique partial index on `(user_id)` where `status = 'active'`

### cycle_settings

Current editable settings for a cycle.

Settings can be changed while progressing through the cycle. Progress checkpoints keep prescribed snapshots, so old completed history does not silently change when settings are edited later.

Progression step rules:

- `step_4_percent` is the default and matches the current XLSX behavior.
- `step_5_percent` is an explicit user preference, not an automatic `variant_2` behavior.
- Variant still controls only `basePercent`: `variant_1 = 0.82`, `variant_2 = 0.86`.

| Column                | Type             | Notes                                |
|-----------------------|------------------|--------------------------------------|
| `cycle_id`            | uuid primary key | FK `program_cycles.id`               |
| `deadlift_1rm_kg`     | numeric          |                                      |
| `bench_1rm_kg`        | numeric          |                                      |
| `squat_1rm_kg`        | numeric          |                                      |
| `variant`             | text             | `variant_1` or `variant_2`           |
| `progression_step`    | text             | `step_4_percent` or `step_5_percent` |
| `deadlift_assistance` | text             | Exercise option id                   |
| `bench_assistance`    | text             | Exercise option id                   |
| `squat_assistance`    | text             | Exercise option id                   |
| `gpp_abs`             | text nullable    | `null` means skipped                 |
| `gpp_triceps`         | text nullable    |                                      |
| `gpp_horizontal_pull` | text nullable    |                                      |
| `gpp_biceps`          | text nullable    |                                      |
| `gpp_vertical_pull`   | text nullable    |                                      |
| `gpp_overhead_press`  | text nullable    |                                      |
| `created_at`          | timestamptz      |                                      |
| `updated_at`          | timestamptz      |                                      |

### cycle_settings_revisions

Optional but recommended audit trail for settings changes.

| Column          | Type          | Notes                               |
|-----------------|---------------|-------------------------------------|
| `id`            | uuid          |                                     |
| `cycle_id`      | uuid          | FK `program_cycles.id`              |
| `settings_json` | jsonb         | Full settings snapshot after change |
| `reason`        | text nullable | User/system note                    |
| `created_at`    | timestamptz   |                                     |

### progress_checkpoints

User progress against calculated training rows.

| Column                 | Type                 | Notes                                   |
|------------------------|----------------------|-----------------------------------------|
| `id`                   | uuid primary key     |                                         |
| `cycle_id`             | uuid                 | FK `program_cycles.id`                  |
| `week`                 | text                 | Week at the time of completion          |
| `variant`              | text                 | Variant at the time of completion       |
| `day_id`               | text                 | `day_1`, `day_2`, `day_3`               |
| `exercise_key`         | text                 | Stable program exercise key             |
| `row_kind`             | text                 | `main`, `assistance`, `gpp`             |
| `status`               | text                 | `planned`, `done`, `skipped`, `partial` |
| `prescribed_sets`      | integer nullable     | Parsed from prescription when possible  |
| `prescribed_reps`      | text nullable        | Keep ranges as text, e.g. `6-10`        |
| `prescribed_weight_kg` | numeric nullable     |                                         |
| `prescribed_rpe`       | text nullable        |                                         |
| `completed_sets`       | integer nullable     |                                         |
| `completed_reps`       | text nullable        | Allows `5,5,4` or ranges                |
| `actual_weight_kg`     | numeric nullable     |                                         |
| `actual_rpe`           | text nullable        |                                         |
| `notes`                | text nullable        |                                         |
| `completed_at`         | timestamptz nullable |                                         |
| `created_at`           | timestamptz          |                                         |
| `updated_at`           | timestamptz          |                                         |

Recommended unique key:

- `cycle_id`
- `week`
- `day_id`
- `exercise_key`

This keeps checkpoint updates idempotent for the current MVP.

### exercise_aliases

Mapping between program exercise names and dataset exercise ids.

| Column                 | Type             | Notes                                  |
|------------------------|------------------|----------------------------------------|
| `program_exercise_key` | text primary key | Stable app key                         |
| `program_name_ru`      | text             | Name from XLSX/UI                      |
| `dataset_exercise_id`  | text nullable    | `null` when no match is confirmed      |
| `dataset_name`         | text nullable    |                                        |
| `review_status`        | text             | `confirmed`, `missing`, `needs_review` |
| `notes`                | text nullable    |                                        |

### exercise_media

Stable media projection for confirmed dataset exercises.

| Column                | Type             | Notes                             |
|-----------------------|------------------|-----------------------------------|
| `dataset_exercise_id` | text primary key |                                   |
| `gif_url`             | text nullable    | Stable URL returned to the client |
| `width`               | integer nullable | For layout reservation            |
| `height`              | integer nullable | For layout reservation            |
| `updated_at`          | timestamptz      |                                   |

## MVP Persistence Rules

- Persist users, sessions, active cycle settings, and progress checkpoints.
- Keep calculation deterministic from cycle settings plus selected week.
- Allow editing cycle settings while the cycle is active.
- Keep prescribed values on checkpoints as snapshots.
- Do not persist derived training rows unless later needed for audit/history.
