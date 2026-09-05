package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"sport/server/internal/api"
	"sport/server/internal/exercises"
	"sport/server/internal/program"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresReferenceData(t *testing.T) {
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
	schema := "sport_test_" + uuid.New().String()[:8]
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
	for i := 0; i < 2; i++ {
		if err = RunMigrations(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	var versions int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&versions); err != nil || versions != 3 {
		t.Fatalf("migrations = %d: %v", versions, err)
	}
	dir := filepath.Join("..", "..", "..", "exercises-dataset-main")
	catalog, err := exercises.NewPostgresCatalog(ctx, pool, dir)
	if err != nil {
		t.Fatal(err)
	}
	all, err := catalog.Search(ctx, api.ListExercisesParams{Limit: api.NewOptInt(10)})
	if err != nil {
		t.Fatal(err)
	}
	if all.Total != 1324 || len(all.Items) != 10 || len(all.Muscles) == 0 || len(all.Equipment) == 0 {
		t.Fatalf("invalid catalog: total=%d, items=%d", all.Total, len(all.Items))
	}
	first := all.Items[0]
	filtered, err := catalog.Search(ctx, api.ListExercisesParams{Muscle: api.NewOptString(first.TargetMuscles[0]), Equipment: api.NewOptString(first.Equipment.Or("")), HasImage: api.NewOptBool(true)})
	if err != nil || filtered.Total == 0 {
		t.Fatalf("filters: %v", err)
	}
	for _, item := range filtered.Items {
		if item.Equipment.Or("") != first.Equipment.Or("") {
			t.Fatal("equipment filter ignored")
		}
	}
	page, err := catalog.Search(ctx, api.ListExercisesParams{Offset: api.NewOptInt(all.Total + 1)})
	if err != nil || page.Total != all.Total || len(page.Items) != 0 {
		t.Fatalf("out of range page: %v", err)
	}
	literal, err := catalog.Search(ctx, api.ListExercisesParams{Query: api.NewOptString("%_' OR 1=1 --")})
	if err != nil || literal.Total != 0 {
		t.Fatalf("query was not literal: %v", err)
	}
	ru, err := catalog.Search(ctx, api.ListExercisesParams{Query: api.NewOptString("Становая тяга")})
	if err != nil || ru.Total == 0 {
		t.Fatalf("Russian search: %v", err)
	}
	details, ok, err := catalog.GetDetails(ctx, "deadlift")
	if err != nil || !ok || details.DatasetExerciseId.Or("") != "0032" {
		t.Fatalf("alias: %v", err)
	}
	if _, ok, err = catalog.Get(ctx, "unknown"); err != nil || ok {
		t.Fatalf("unknown exercise: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE exercise_catalog SET name='Edited in DB' WHERE id='0032'`); err != nil {
		t.Fatal(err)
	}
	// Subsequent startup must not need JSON or overwrite database edits.
	catalog, err = exercises.NewPostgresCatalog(ctx, pool, "/missing-dataset")
	if err != nil {
		t.Fatal(err)
	}
	edited, ok, err := catalog.Get(ctx, "0032")
	if err != nil || !ok || edited.Name != "Edited in DB" {
		t.Fatalf("seed overwrote DB: %v", err)
	}
	store := NewPostgresStore(pool)
	options, err := store.ProgramOptions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Verify that SQL seeds support the actual default program and its known result.
	baseline, err := program.NewCalculator(options).Calculate(api.ProgramSelection{Settings: program.DefaultSettings(), Week: api.ProgramWeekWeek1})
	if err != nil {
		t.Fatalf("SQL program options reject default settings: %v", err)
	}
	if len(options.Weeks) != 8 || len(baseline.Days) != 3 || baseline.Days[0].Rows[0].Prescription.WeightKg.Or(0) != 145 {
		t.Fatal("SQL program seed does not produce the expected baseline plan")
	}
	options.Assistance.Deadlift = []api.SelectOption{{ID: "new_lift", Label: "Новая тяга"}}
	data, _ := json.Marshal(options)
	if _, err = pool.Exec(ctx, `UPDATE program_options SET options=$1 WHERE id=$2`, data, program.FormulaVersion); err != nil {
		t.Fatal(err)
	}
	settings := program.DefaultSettings()
	settings.Assistance.Deadlift = "new_lift"
	plan, err := NewHandler(store, catalog).calculate(ctx, api.ProgramSelection{Settings: settings, Week: api.ProgramWeekWeek1})
	if err != nil || plan.Days[0].Rows[2].ExerciseName != "Новая тяга" {
		t.Fatalf("calculator did not read DB options: %v", err)
	}
}
