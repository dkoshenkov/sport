package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"sport/server/internal/api"
	"sport/server/internal/exercises"
)

func TestMigrateLegacyCatalog(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := "sport_legacy_" + uuid.New().String()[:8]
	if _, err = admin.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	// Column names and nullability from the deployed schema supplied by the user.
	_, err = pool.Exec(ctx, `
 CREATE TABLE exercise_catalog (
 dataset_exercise_id text PRIMARY KEY, name text NOT NULL,
 category text, body_part text, equipment text, target text, muscle_group text,
 secondary_muscles text[] NOT NULL DEFAULT '{}',
 instructions jsonb NOT NULL DEFAULT '{}', instruction_steps jsonb NOT NULL DEFAULT '{}',
 image_path text, gif_path text, source_created_at text,
 imported_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now());
 CREATE TABLE exercise_aliases (
 program_exercise_key text PRIMARY KEY, program_name_ru text NOT NULL,
 dataset_exercise_id text REFERENCES exercise_catalog(dataset_exercise_id) ON DELETE SET NULL,
 dataset_name text, review_status text NOT NULL, notes text,
 name_hints text[] NOT NULL DEFAULT '{}', updated_at timestamptz NOT NULL DEFAULT now());
 CREATE TABLE exercise_media (
 dataset_exercise_id text REFERENCES exercise_catalog(dataset_exercise_id) ON DELETE CASCADE, payload text);
 CREATE TABLE exercise_translations_ru (
 dataset_exercise_id text REFERENCES exercise_catalog(dataset_exercise_id) ON DELETE CASCADE, payload text);
 INSERT INTO exercise_catalog(dataset_exercise_id,name,gif_path,muscle_group,instructions,instruction_steps)
 VALUES ('0032','Custom deadlift','videos/0032-ila4NZS.gif','Legacy group','{"ru":["Legacy text"]}','{"ru":["Custom technique"]}'),
 ('0025','Custom bench',NULL,NULL,'{}','{}');
 INSERT INTO exercise_aliases(program_exercise_key,program_name_ru,dataset_exercise_id,review_status,notes)
 VALUES ('deadlift','Моя становая','0032','confirmed','Keep this note');
 INSERT INTO exercise_media VALUES ('0032','Keep media');
 INSERT INTO exercise_translations_ru VALUES ('0032','Keep translation');`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = RunMigrations(ctx, pool); err != nil {
			t.Fatalf("legacy migration: %v", err)
		}
	}
	catalog, err := exercises.NewPostgresCatalog(ctx, pool, filepath.Join("..", "..", "..", "exercises-dataset-main"))
	if err != nil {
		t.Fatal(err)
	}
	item, ok, err := catalog.Get(ctx, "0032")
	if err != nil || !ok {
		t.Fatalf("read migrated row: %v", err)
	}
	if item.Name != "Custom deadlift" || len(item.Instructions) != 1 || item.Instructions[0] != "Custom technique" || item.Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("legacy data changed or media missing: %#v", item)
	}
	details, ok, err := catalog.GetDetails(ctx, "deadlift")
	if err != nil || !ok || details.Name != "Моя становая" || details.DatasetExerciseId.Or("") != "0032" {
		t.Fatalf("legacy alias: %#v, %v", details, err)
	}
	page, err := catalog.Search(ctx, api.ListExercisesParams{Query: api.NewOptString("МОЯ СТАНОВАЯ"), HasImage: api.NewOptBool(true)})
	if err != nil || page.Total != 1 {
		t.Fatalf("legacy search/media filter: total=%d, %v", page.Total, err)
	}
	bench, ok, err := catalog.Get(ctx, "0025")
	if err != nil || !ok || bench.Name != "Custom bench" || bench.Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("missing legacy path fallback: %#v, %v", bench, err)
	}
	var preserved bool
	err = pool.QueryRow(ctx, `SELECT c.muscle_group='Legacy group' AND c.instructions='{"ru":["Legacy text"]}'::jsonb AND c.gif_path='videos/0032-ila4NZS.gif' AND a.notes='Keep this note' AND m.payload='Keep media' AND t.payload='Keep translation'
 FROM exercise_catalog c JOIN exercise_aliases a ON a.dataset_id=c.id JOIN exercise_media m ON m.dataset_exercise_id=c.id JOIN exercise_translations_ru t ON t.dataset_exercise_id=c.id WHERE c.id='0032'`).Scan(&preserved)
	if err != nil || !preserved {
		t.Fatalf("legacy fields or related rows lost: %v", err)
	}
	// Verify the incoming FK actions still work after renaming the referenced PK.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM exercise_catalog WHERE id='0032'`); err != nil {
		t.Fatal(err)
	}
	err = tx.QueryRow(ctx, `SELECT (SELECT dataset_id IS NULL FROM exercise_aliases WHERE program_key='deadlift') AND NOT EXISTS(SELECT 1 FROM exercise_media) AND NOT EXISTS(SELECT 1 FROM exercise_translations_ru)`).Scan(&preserved)
	if err != nil || !preserved {
		t.Fatalf("foreign key actions changed: %v", err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = exercises.NewPostgresCatalog(ctx, pool, "/missing-dataset"); err != nil {
		t.Fatalf("repeat startup: %v", err)
	}
}
