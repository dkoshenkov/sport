package exercises

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sport/server/internal/api"
)

type Repository interface {
	Search(context.Context, api.ListExercisesParams) (api.ExerciseCatalogListResponse, error)
	Get(context.Context, string) (api.ExerciseCatalogItem, bool, error)
	GetDetails(context.Context, string) (api.ExerciseDetails, bool, error)
}

type PostgresCatalog struct {
	pool         *pgxpool.Pool
	mediaCatalog *Catalog
}

func NewPostgresCatalog(ctx context.Context, pool *pgxpool.Pool, datasetDir string) (*PostgresCatalog, error) {
	c := &PostgresCatalog{pool: pool, mediaCatalog: &Catalog{datasetDir: datasetDir, mediaBaseURL: "/videos"}}
	if err := c.seed(ctx, datasetDir); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *PostgresCatalog) seed(ctx context.Context, dir string) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(74219002)`); err != nil {
		return err
	}
	var imported bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_imports WHERE name='exercise-dataset-v1')`).Scan(&imported); err != nil {
		return err
	}
	if imported {
		return tx.Commit(ctx)
	}
	source, err := NewCatalog(dir)
	if err != nil {
		return err
	}
	for _, item := range source.exercises {
		instructions, err := json.Marshal(item.InstructionSteps)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO exercise_catalog(id,name,category,body_part,equipment,target,secondary_muscles,instruction_steps,gif_url,has_media)
   VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(id) DO NOTHING`, item.ID, item.Name, item.Category, item.BodyPart, item.Equipment, item.Target, nonEmpty(item.SecondaryMuscles), instructions, item.GIFURL, source.hasMedia(item))
		if err != nil {
			return fmt.Errorf("import exercise %s: %w", item.ID, err)
		}
	}
	// Existing legacy rows were not inserted above. Initialize their media metadata
	// from their preserved paths; only fill an absent path from the source dataset.
	rows, err := tx.Query(ctx, `SELECT id, gif_url FROM exercise_catalog`)
	if err != nil {
		return err
	}
	var mediaItems []datasetExercise
	for rows.Next() {
		var item datasetExercise
		if err = rows.Scan(&item.ID, &item.GIFURL); err != nil {
			rows.Close()
			return err
		}
		if item.GIFURL == "" {
			item.GIFURL = source.byDatasetID[item.ID].GIFURL
		}
		mediaItems = append(mediaItems, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}
	for _, item := range mediaItems {
		if _, err = tx.Exec(ctx, `UPDATE exercise_catalog SET gif_url=$2, has_media=$3 WHERE id=$1`, item.ID, item.GIFURL, source.hasMedia(item)); err != nil {
			return err
		}
	}
	// Resolve imported program references once. Keep unresolved variants explicit.
	if _, err = tx.Exec(ctx, `UPDATE exercise_aliases a SET dataset_id=(
 SELECT e.id FROM exercise_catalog e
 WHERE e.id=ANY(a.name_hints) OR EXISTS(
 SELECT 1 FROM unnest(a.name_hints) hint
 WHERE regexp_replace(lower(replace(e.name,'-',' ')), '\s+', ' ', 'g')=regexp_replace(lower(replace(hint,'-',' ')), '\s+', ' ', 'g'))
 ORDER BY (e.id=ANY(a.name_hints)) DESC,e.id LIMIT 1)
 WHERE a.dataset_id IS NULL AND a.review_status<>'missing'`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO reference_imports(name) VALUES('exercise-dataset-v1')`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const catalogColumns = `e.id,e.name,COALESCE(e.category,''),COALESCE(e.body_part,''),COALESCE(e.equipment,''),COALESCE(e.target,''),e.secondary_muscles,e.instruction_steps,e.gif_url,
 COALESCE((SELECT a.program_name FROM exercise_aliases a WHERE a.dataset_id=e.id ORDER BY a.program_key LIMIT 1),'')`

func (c *PostgresCatalog) scan(row pgx.Row) (api.ExerciseCatalogItem, error) {
	var item datasetExercise
	var instructions []byte
	var nameRu string
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.BodyPart, &item.Equipment, &item.Target, &item.SecondaryMuscles, &instructions, &item.GIFURL, &nameRu)
	if err != nil {
		return api.ExerciseCatalogItem{}, err
	}
	if err = json.Unmarshal(instructions, &item.InstructionSteps); err != nil {
		return api.ExerciseCatalogItem{}, err
	}
	result := c.mediaCatalog.catalogItem(item)
	if nameRu != "" {
		result.NameRu = api.NewOptNilString(nameRu)
	}
	return result, nil
}

func (c *PostgresCatalog) Search(ctx context.Context, params api.ListExercisesParams) (api.ExerciseCatalogListResponse, error) {
	limit, offset := params.Limit.Or(30), params.Offset.Or(0)
	if limit < 1 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	result := api.ExerciseCatalogListResponse{Items: []api.ExerciseCatalogItem{}, Limit: limit, Offset: offset}
	query := strings.ToLower(strings.TrimSpace(params.Query.Or("")))
	// strpos treats user input literally, including SQL wildcard characters.
	where := ` FROM exercise_catalog e WHERE ($1='' OR strpos(lower(concat_ws(' ',e.id,e.name,e.category,e.body_part,e.equipment,e.target,array_to_string(e.secondary_muscles,' ')) COLLATE sport_unicode),$1)>0
 OR EXISTS(SELECT 1 FROM exercise_aliases a WHERE a.dataset_id=e.id AND strpos(lower(concat_ws(' ',a.program_key,a.program_name,array_to_string(a.name_hints,' ')) COLLATE sport_unicode),$1)>0))
 AND (NOT $2 OR e.has_media) AND ($3='' OR e.target=$3 OR $3=ANY(e.secondary_muscles)) AND ($4='' OR e.equipment=$4) AND ($5='' OR e.body_part=$5)`
	args := []any{query, params.HasImage.Or(false), params.Muscle.Or(""), params.Equipment.Or(""), params.BodyPart.Or("")}
	// A shared snapshot keeps total and page consistent while catalog data is edited.
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	if err = tx.QueryRow(ctx, `SELECT count(*)`+where, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT `+catalogColumns+where+` ORDER BY lower(e.name),e.id LIMIT $6 OFFSET $7`, append(args, limit, offset)...)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		item, err := c.scan(rows)
		if err != nil {
			rows.Close()
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, err
	}
	// Facets come from the entire catalog, never just the current page.
	if err = tx.QueryRow(ctx, `SELECT
 ARRAY(SELECT DISTINCT muscle FROM (SELECT target AS muscle FROM exercise_catalog UNION SELECT unnest(secondary_muscles) FROM exercise_catalog) m WHERE muscle<>'' ORDER BY muscle),
 ARRAY(SELECT DISTINCT equipment FROM exercise_catalog WHERE equipment<>'' ORDER BY equipment),
 ARRAY(SELECT DISTINCT body_part FROM exercise_catalog WHERE body_part<>'' ORDER BY body_part)`).Scan(&result.Muscles, &result.Equipment, &result.BodyParts); err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}

func (c *PostgresCatalog) Get(ctx context.Context, id string) (api.ExerciseCatalogItem, bool, error) {
	item, err := c.scan(c.pool.QueryRow(ctx, `SELECT `+catalogColumns+` FROM exercise_catalog e WHERE e.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return item, false, nil
	}
	return item, err == nil, err
}

func (c *PostgresCatalog) GetDetails(ctx context.Context, key string) (api.ExerciseDetails, bool, error) {
	var name, status string
	var id *string
	err := c.pool.QueryRow(ctx, `SELECT program_name,review_status,dataset_id FROM exercise_aliases WHERE program_key=$1`, key).Scan(&name, &status, &id)
	if errors.Is(err, pgx.ErrNoRows) {
		return api.ExerciseDetails{}, false, nil
	}
	if err != nil {
		return api.ExerciseDetails{}, false, err
	}
	result := api.ExerciseDetails{ExerciseKey: key, Name: name, AliasStatus: api.ExerciseDetailsAliasStatus(status), TargetMuscles: []string{}, SecondaryMuscles: []string{}, Instructions: []string{}, Media: api.ExerciseMedia{Status: api.ExerciseMediaStatusMissing}}
	if id != nil {
		item, ok, err := c.Get(ctx, *id)
		if err != nil {
			return result, false, err
		}
		if ok {
			result.DatasetExerciseId = api.NewOptNilString(item.DatasetExerciseId)
			result.DatasetName = api.NewOptNilString(item.Name)
			result.Equipment = item.Equipment
			result.TargetMuscles = item.TargetMuscles
			result.SecondaryMuscles = item.SecondaryMuscles
			result.Instructions = item.Instructions
			result.Media = item.Media
		}
	}
	return result, true, nil
}

// In-memory adapter for unit tests and seed validation.
func (c *Catalog) Search(_ context.Context, p api.ListExercisesParams) (api.ExerciseCatalogListResponse, error) {
	return c.List(p), nil
}
func (c *Catalog) Get(_ context.Context, id string) (api.ExerciseCatalogItem, bool, error) {
	item, ok := c.CatalogExercise(id)
	return item, ok, nil
}
func (c *Catalog) GetDetails(_ context.Context, key string) (api.ExerciseDetails, bool, error) {
	item, ok := c.Details(key)
	return item, ok, nil
}
