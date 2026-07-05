package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dkoshenkov/packages-go/configx"
	"github.com/dkoshenkov/packages-go/logx"
	"github.com/jackc/pgx/v5/pgxpool"

	"sport/server/internal/app"
	"sport/server/internal/exercises"
)

type syncConfig struct {
	ServiceName string             `cfgx:"service_name,default=sport-media-sync" env:"SERVICE_NAME"`
	Env         string             `cfgx:"env,default=dev" env:"ENV"`
	Log         syncLogConfig      `cfgx:"log"`
	RepoURL     string             `cfgx:"repo_url,default=https://github.com/hasaneyldrm/exercises-dataset.git" env:"EXERCISE_DATASET_REPO_URL"`
	DatasetDir  string             `cfgx:"dataset_dir,optional" env:"EXERCISE_DATASET_DIR"`
	Out         string             `cfgx:"out,default=exercise-media.json" env:"EXERCISE_MEDIA_MANIFEST"`
	KeepTemp    bool               `cfgx:"keep_temp,default=false" env:"KEEP_TEMP"`
	RUOverrides string             `cfgx:"ru_overrides,default=internal/exercises/ru_overrides.json" env:"EXERCISE_RU_OVERRIDES"`
	Database    syncDatabaseConfig `cfgx:"database"`
	S3          syncS3Config       `cfgx:"s3"`
	Media       syncMediaConfig    `cfgx:"media"`
	Storage     syncStorageConfig  `cfgx:"storage"`
}

type syncLogConfig struct {
	Level  string `cfgx:"level,default=info" env:"LOG_LEVEL"`
	Pretty bool   `cfgx:"pretty,default=false" env:"LOG_PRETTY"`
}

type syncDatabaseConfig struct {
	URL           string `cfgx:"url,default=postgres://sport:sport@localhost:5432/sport?sslmode=disable" env:"DATABASE_URL"`
	RunMigrations bool   `cfgx:"run_migrations,default=true" env:"DATABASE_RUN_MIGRATIONS"`
}

type syncS3Config struct {
	Bucket string `cfgx:"bucket,optional" env:"S3_BUCKET"`
	Prefix string `cfgx:"prefix,default=exercises" env:"S3_PREFIX"`
}

type syncMediaConfig struct {
	MainDomain string `cfgx:"main_domain,default=example.com" env:"SPORT_MAIN_DOMAIN"`
	BaseURL    string `cfgx:"base_url,optional" env:"EXERCISE_MEDIA_BASE_URL"`
}

type syncStorageConfig struct {
	Mode          string `cfgx:"mode,default=local" env:"EXERCISE_MEDIA_STORAGE_MODE"`
	LocalDir      string `cfgx:"local_dir,default=../client/public/exercises" env:"EXERCISE_MEDIA_LOCAL_DIR"`
	SourceDir     string `cfgx:"source_dir,optional" env:"EXERCISE_MEDIA_SOURCE_DIR"`
	SourceBaseURL string `cfgx:"source_base_url,optional" env:"EXERCISE_MEDIA_SOURCE_BASE_URL"`
}

type datasetExercise struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Category         string              `json:"category"`
	BodyPart         string              `json:"body_part"`
	Equipment        string              `json:"equipment"`
	Target           string              `json:"target"`
	MuscleGroup      string              `json:"muscle_group"`
	SecondaryMuscles []string            `json:"secondary_muscles"`
	Instructions     map[string]string   `json:"instructions"`
	InstructionSteps map[string][]string `json:"instruction_steps"`
	Image            string              `json:"image"`
	GIFURL           string              `json:"gif_url"`
	MediaID          string              `json:"media_id"`
	CreatedAt        string              `json:"created_at"`
}

type ruOverride struct {
	ProgramKey   string   `json:"programExerciseKey"`
	DatasetID    string   `json:"datasetExerciseId"`
	Name         string   `json:"nameRu"`
	Instructions []string `json:"instructionsRu"`
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "sync exercise media failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadSyncConfig(ctx)
	if err != nil {
		return err
	}
	logger, err := logx.New(cfg.ServiceName, logOptions(cfg)...)
	if err != nil {
		return err
	}
	ctx = logx.WithContext(ctx, logger)
	if err := validateConfig(cfg); err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "exercises-dataset-*")
	if err != nil {
		return err
	}
	if !cfg.KeepTemp {
		defer os.RemoveAll(tmp)
	}
	repoDir, err := prepareDatasetDir(ctx, cfg, tmp)
	if err != nil {
		return err
	}

	list, byID, err := readDataset(repoDir)
	if err != nil {
		return err
	}
	overrides, err := readRUOverrides(cfg.RUOverrides)
	if err != nil {
		return err
	}

	var pool *pgxpool.Pool
	if cfg.Database.URL != "" {
		pool, err = pgxpool.New(ctx, cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("connect postgres: %w", err)
		}
		defer pool.Close()
		if err := pool.Ping(ctx); err != nil {
			return fmt.Errorf("ping postgres: %w", err)
		}
		if cfg.Database.RunMigrations {
			if err := app.RunMigrations(ctx, pool); err != nil {
				return err
			}
		}
		if err := upsertCatalog(ctx, pool, list); err != nil {
			return err
		}
		if err := upsertAliases(ctx, pool, byID); err != nil {
			return err
		}
		if err := upsertRUOverrides(ctx, pool, byID, overrides); err != nil {
			return err
		}
	}

	s3Client, err := newS3Client(ctx, cfg)
	if err != nil {
		return err
	}
	manifest := make([]exercises.Media, 0, len(list))
	for _, item := range list {
		media, err := syncMedia(ctx, cfg, s3Client, repoDir, item)
		if err != nil {
			return fmt.Errorf("%s: sync media: %w", item.ID, err)
		}
		manifest = append(manifest, media)
		if pool != nil {
			if err := upsertMedia(ctx, pool, media); err != nil {
				return err
			}
		}
		if media.GIFURL == "" {
			logx.Warn(ctx).Str("dataset_id", item.ID).Str("gif_path", item.GIFURL).Msg("exercise gif missing")
			continue
		}
		logx.Info(ctx).Str("dataset_id", item.ID).Str("url", media.GIFURL).Msg("synced exercise gif")
	}
	return writeManifest(cfg.Out, manifest)
}

func loadSyncConfig(ctx context.Context) (syncConfig, error) {
	var cfg syncConfig
	profile := os.Getenv("ENV")
	if profile == "" {
		profile = "dev"
	}
	err := configx.Load(ctx, &cfg,
		configx.ParseFlags(),
		configx.WithProfile(profile),
		configx.WithResolveMode(configx.OverlayDefaultLow),
	)
	if cfg.Media.BaseURL == "" && cfg.Media.MainDomain != "" && cfg.Storage.Mode == "s3" {
		cfg.Media.BaseURL = "https://media." + cfg.Media.MainDomain
	}
	return cfg, err
}

func validateConfig(cfg syncConfig) error {
	switch cfg.Storage.Mode {
	case "none", "local", "s3":
	default:
		return fmt.Errorf("unsupported storage mode %q", cfg.Storage.Mode)
	}
	if strings.TrimSpace(cfg.Out) == "" {
		return errors.New("exercise media manifest path is required")
	}
	if cfg.Storage.Mode == "s3" && cfg.S3.Bucket == "" {
		return errors.New("S3 bucket is required when EXERCISE_MEDIA_STORAGE_MODE=s3")
	}
	if cfg.Storage.Mode == "s3" && cfg.Media.BaseURL == "" {
		return errors.New("public base URL is required: set EXERCISE_MEDIA_BASE_URL or SPORT_MAIN_DOMAIN")
	}
	return nil
}

func logOptions(cfg syncConfig) []logx.Option {
	opts := []logx.Option{
		logx.WithLevelText(cfg.Log.Level),
		logx.WithField("env", cfg.Env),
	}
	if cfg.Log.Pretty {
		opts = append(opts, logx.WithPretty())
	}
	return opts
}

func cloneDataset(ctx context.Context, repoURL, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func prepareDatasetDir(ctx context.Context, cfg syncConfig, tmp string) (string, error) {
	if cfg.DatasetDir != "" {
		info, err := os.Stat(cfg.DatasetDir)
		if err != nil {
			return "", fmt.Errorf("stat dataset dir: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("dataset dir is not a directory: %s", cfg.DatasetDir)
		}
		return cfg.DatasetDir, nil
	}
	repoDir := filepath.Join(tmp, "repo")
	if err := cloneDataset(ctx, cfg.RepoURL, repoDir); err != nil {
		return "", err
	}
	return repoDir, nil
}

func readDataset(repoDir string) ([]datasetExercise, map[string]datasetExercise, error) {
	candidates := []string{
		filepath.Join(repoDir, "data", "exercises.json"),
		filepath.Join(repoDir, "exercises.json"),
	}
	var data []byte
	var err error
	for _, candidate := range candidates {
		data, err = os.ReadFile(candidate)
		if err == nil {
			break
		}
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("dataset JSON not found: %w", err)
	}
	var list []datasetExercise
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	byID := make(map[string]datasetExercise, len(list))
	for _, exercise := range list {
		byID[exercise.ID] = exercise
	}
	return list, byID, nil
}

func readRUOverrides(file string) ([]ruOverride, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var overrides []ruOverride
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, err
	}
	return overrides, nil
}

func newS3Client(ctx context.Context, cfg syncConfig) (*s3.Client, error) {
	if cfg.Storage.Mode != "s3" {
		return nil, nil
	}
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(awsCfg), nil
}

func syncMedia(ctx context.Context, cfg syncConfig, s3Client *s3.Client, repoDir string, exercise datasetExercise) (exercises.Media, error) {
	media := exercises.Media{
		DatasetID:        exercise.ID,
		Equipment:        exercise.Equipment,
		TargetMuscles:    nonEmpty([]string{exercise.Target}),
		SecondaryMuscles: append([]string(nil), exercise.SecondaryMuscles...),
		Instructions:     preferredInstructions(exercise),
		UpdatedAt:        time.Now().UTC().Format(time.RFC3339),
		Provenance:       "missing",
	}
	if cfg.Storage.Mode == "none" {
		return media, nil
	}
	gif, ok, err := loadGIF(ctx, cfg, repoDir, exercise)
	if err != nil || !ok {
		return media, err
	}
	key := strings.Trim(cfg.S3.Prefix, "/") + "/" + exercise.ID + ".gif"
	switch cfg.Storage.Mode {
	case "local":
		if err := os.MkdirAll(cfg.Storage.LocalDir, 0o755); err != nil {
			return exercises.Media{}, err
		}
		if err := os.WriteFile(filepath.Join(cfg.Storage.LocalDir, exercise.ID+".gif"), gif, 0o644); err != nil {
			return exercises.Media{}, err
		}
	case "s3":
		if _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(cfg.S3.Bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(gif),
			ContentType: aws.String("image/gif"),
		}); err != nil {
			return exercises.Media{}, err
		}
	}
	media.GIFURL = publicMediaURL(cfg.Media.BaseURL, key)
	media.StorageKey = key
	media.Provenance = cfg.Storage.Mode
	return media, nil
}

func publicMediaURL(baseURL, key string) string {
	key = strings.TrimLeft(key, "/")
	if strings.TrimSpace(baseURL) == "" {
		return "/" + key
	}
	return strings.TrimRight(baseURL, "/") + "/" + key
}

func loadGIF(ctx context.Context, cfg syncConfig, repoDir string, exercise datasetExercise) ([]byte, bool, error) {
	if u, err := url.Parse(exercise.GIFURL); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		data, err := fetchGIF(ctx, exercise.GIFURL)
		return data, err == nil, err
	}

	if cfg.Storage.SourceDir != "" {
		for _, candidate := range localGIFCandidates(cfg.Storage.SourceDir, exercise) {
			if data, err := os.ReadFile(candidate); err == nil {
				return data, true, nil
			}
		}
	}

	sourceBaseURL := strings.TrimRight(cfg.Storage.SourceBaseURL, "/")
	if sourceBaseURL != "" {
		if mediaID := mediaIDFromExercise(exercise); mediaID != "" {
			data, err := fetchGIF(ctx, sourceBaseURL+"/"+mediaID+".gif")
			if err == nil {
				return data, true, nil
			}
		}
	}

	if exercise.GIFURL != "" {
		if data, err := os.ReadFile(filepath.Join(repoDir, filepath.FromSlash(exercise.GIFURL))); err == nil {
			return data, true, nil
		}
		if sourceBaseURL != "" {
			data, err := fetchGIF(ctx, sourceBaseURL+"/"+strings.TrimLeft(exercise.GIFURL, "/"))
			if err == nil {
				return data, true, nil
			}
		}
	}
	for _, candidate := range []string{
		filepath.Join(repoDir, "exercises", exercise.ID+".gif"),
		filepath.Join(repoDir, "gifs", exercise.ID+".gif"),
		filepath.Join(repoDir, "videos", exercise.ID+".gif"),
	} {
		if data, err := os.ReadFile(candidate); err == nil {
			return data, true, nil
		}
	}
	return nil, false, nil
}

func localGIFCandidates(sourceDir string, exercise datasetExercise) []string {
	names := []string{exercise.ID + ".gif"}
	if mediaID := mediaIDFromExercise(exercise); mediaID != "" {
		names = append(names, mediaID+".gif")
	}
	if exercise.GIFURL != "" {
		names = append(names, filepath.Base(exercise.GIFURL))
	}

	candidates := make([]string, 0, len(names)*2)
	seen := make(map[string]struct{}, len(names)*2)
	for _, name := range names {
		for _, path := range []string{
			filepath.Join(sourceDir, name),
			filepath.Join(sourceDir, "assets", name),
		} {
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			candidates = append(candidates, path)
		}
	}
	return candidates
}

func mediaIDFromExercise(exercise datasetExercise) string {
	if mediaID := strings.TrimSpace(exercise.MediaID); mediaID != "" {
		return mediaID
	}
	if mediaID := mediaIDFromGIFPath(exercise.GIFURL); mediaID != "" {
		return mediaID
	}
	return mediaIDFromImagePath(exercise.Image)
}

func mediaIDFromGIFPath(gifPath string) string {
	if strings.TrimSpace(gifPath) == "" {
		return ""
	}
	base := strings.TrimSuffix(filepath.Base(gifPath), filepath.Ext(gifPath))
	if base == "" {
		return ""
	}
	_, mediaID, ok := strings.Cut(base, "-")
	if !ok {
		return base
	}
	return mediaID
}

func mediaIDFromImagePath(imagePath string) string {
	if strings.TrimSpace(imagePath) == "" {
		return ""
	}
	base := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))
	if base == "" {
		return ""
	}
	_, mediaID, ok := strings.Cut(base, "-")
	if !ok {
		return ""
	}
	return mediaID
}

func fetchGIF(ctx context.Context, gifURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gifURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected gif status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 25<<20))
}

func preferredInstructions(exercise datasetExercise) []string {
	for _, lang := range []string{"ru", "en"} {
		if steps := exercise.InstructionSteps[lang]; len(steps) > 0 {
			return append([]string(nil), steps...)
		}
	}
	for _, steps := range exercise.InstructionSteps {
		if len(steps) > 0 {
			return append([]string(nil), steps...)
		}
	}
	return nil
}

func resolveExercise(alias exercises.Alias, dataset map[string]datasetExercise) (datasetExercise, bool) {
	if alias.DatasetID != "" {
		exercise, ok := dataset[alias.DatasetID]
		return exercise, ok
	}
	if alias.Status == exercises.StatusMissing {
		return datasetExercise{}, false
	}
	hints := normalizedSet(alias.NameHints)
	for _, exercise := range dataset {
		if _, ok := hints[normalize(exercise.Name)]; ok {
			return exercise, true
		}
	}
	return datasetExercise{}, false
}

func upsertCatalog(ctx context.Context, pool *pgxpool.Pool, list []datasetExercise) error {
	for _, item := range list {
		instructions, err := json.Marshal(item.Instructions)
		if err != nil {
			return err
		}
		steps, err := json.Marshal(item.InstructionSteps)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO exercise_catalog (
				dataset_exercise_id, name, category, body_part, equipment, target, muscle_group,
				secondary_muscles, instructions, instruction_steps, image_path, gif_path, source_created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11, $12, $13)
			ON CONFLICT (dataset_exercise_id) DO UPDATE SET
				name = EXCLUDED.name,
				category = EXCLUDED.category,
				body_part = EXCLUDED.body_part,
				equipment = EXCLUDED.equipment,
				target = EXCLUDED.target,
				muscle_group = EXCLUDED.muscle_group,
				secondary_muscles = EXCLUDED.secondary_muscles,
				instructions = EXCLUDED.instructions,
				instruction_steps = EXCLUDED.instruction_steps,
				image_path = EXCLUDED.image_path,
				gif_path = EXCLUDED.gif_path,
				source_created_at = EXCLUDED.source_created_at,
				updated_at = now()
		`, item.ID, item.Name, item.Category, item.BodyPart, item.Equipment, item.Target, item.MuscleGroup,
			nonNilStrings(item.SecondaryMuscles), string(instructions), string(steps), item.Image, item.GIFURL, item.CreatedAt); err != nil {
			return fmt.Errorf("%s: upsert catalog: %w", item.ID, err)
		}
	}
	return nil
}

func upsertAliases(ctx context.Context, pool *pgxpool.Pool, dataset map[string]datasetExercise) error {
	for _, alias := range exercises.Aliases() {
		exercise, ok := resolveExercise(alias, dataset)
		datasetID := alias.DatasetID
		datasetName := alias.DatasetName
		if ok {
			datasetID = exercise.ID
			datasetName = exercise.Name
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO exercise_aliases (
				program_exercise_key, program_name_ru, dataset_exercise_id, dataset_name,
				review_status, notes, name_hints
			)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, NULLIF($6, ''), $7)
			ON CONFLICT (program_exercise_key) DO UPDATE SET
				program_name_ru = EXCLUDED.program_name_ru,
				dataset_exercise_id = EXCLUDED.dataset_exercise_id,
				dataset_name = EXCLUDED.dataset_name,
				review_status = EXCLUDED.review_status,
				notes = EXCLUDED.notes,
				name_hints = EXCLUDED.name_hints,
				updated_at = now()
		`, alias.ProgramKey, alias.ProgramName, datasetID, datasetName, alias.Status, alias.Notes, nonNilStrings(alias.NameHints)); err != nil {
			return fmt.Errorf("%s: upsert alias: %w", alias.ProgramKey, err)
		}
	}
	return nil
}

func upsertRUOverrides(ctx context.Context, pool *pgxpool.Pool, dataset map[string]datasetExercise, overrides []ruOverride) error {
	aliases := make(map[string]exercises.Alias)
	for _, alias := range exercises.Aliases() {
		aliases[alias.ProgramKey] = alias
	}
	for _, override := range overrides {
		datasetID := override.DatasetID
		if datasetID == "" && override.ProgramKey != "" {
			if alias, ok := aliases[override.ProgramKey]; ok {
				if exercise, resolved := resolveExercise(alias, dataset); resolved {
					datasetID = exercise.ID
				}
			}
		}
		if datasetID == "" {
			continue
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO exercise_translations_ru (dataset_exercise_id, name_ru, instructions_ru)
			VALUES ($1, NULLIF($2, ''), $3)
			ON CONFLICT (dataset_exercise_id) DO UPDATE SET
				name_ru = EXCLUDED.name_ru,
				instructions_ru = EXCLUDED.instructions_ru,
				updated_at = now()
		`, datasetID, override.Name, nonNilStrings(override.Instructions)); err != nil {
			return fmt.Errorf("%s: upsert ru override: %w", datasetID, err)
		}
	}
	return nil
}

func upsertMedia(ctx context.Context, pool *pgxpool.Pool, media exercises.Media) error {
	status := "missing"
	if media.GIFURL != "" {
		status = "available"
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO exercise_media (dataset_exercise_id, status, gif_url, storage_key, width, height, provenance, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, 0), NULLIF($6, 0), NULLIF($7, ''), now())
		ON CONFLICT (dataset_exercise_id) DO UPDATE SET
			status = EXCLUDED.status,
			gif_url = EXCLUDED.gif_url,
			storage_key = EXCLUDED.storage_key,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			provenance = EXCLUDED.provenance,
			updated_at = now()
	`, media.DatasetID, status, media.GIFURL, media.StorageKey, media.Width, media.Height, media.Provenance)
	if err != nil {
		return fmt.Errorf("%s: upsert media: %w", media.DatasetID, err)
	}
	return nil
}

func writeManifest(file string, manifest []exercises.Media) error {
	if dir := filepath.Dir(file); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(file, data, 0o644)
}

func normalizedSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[normalize(value)] = struct{}{}
	}
	return out
}

func normalize(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(value, "-", " "))), " ")
}

func nonEmpty(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
