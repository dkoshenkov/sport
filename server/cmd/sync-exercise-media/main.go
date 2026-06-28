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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dkoshenkov/packages-go/configx"
	"github.com/dkoshenkov/packages-go/logx"

	"sport/server/internal/exercises"
)

type syncConfig struct {
	ServiceName string          `cfgx:"service_name,default=sport-media-sync" env:"SERVICE_NAME"`
	Env         string          `cfgx:"env,default=dev" env:"ENV"`
	Log         syncLogConfig   `cfgx:"log"`
	RepoURL     string          `cfgx:"repo_url,default=https://github.com/hasaneyldrm/exercises-dataset.git" env:"EXERCISE_DATASET_REPO_URL"`
	Out         string          `cfgx:"out,default=exercise-media.json" env:"EXERCISE_MEDIA_MANIFEST"`
	KeepTemp    bool            `cfgx:"keep_temp,default=false" env:"KEEP_TEMP"`
	S3          syncS3Config    `cfgx:"s3"`
	Media       syncMediaConfig `cfgx:"media"`
}

type syncLogConfig struct {
	Level  string `cfgx:"level,default=info" env:"LOG_LEVEL"`
	Pretty bool   `cfgx:"pretty,default=false" env:"LOG_PRETTY"`
}

type syncS3Config struct {
	Bucket string `cfgx:"bucket" env:"S3_BUCKET"`
	Prefix string `cfgx:"prefix,default=exercises" env:"S3_PREFIX"`
}

type syncMediaConfig struct {
	MainDomain string `cfgx:"main_domain,default=example.com" env:"SPORT_MAIN_DOMAIN"`
	BaseURL    string `cfgx:"base_url,optional" env:"EXERCISE_MEDIA_BASE_URL"`
}

type datasetExercise struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	GIFURL           string   `json:"gifUrl"`
	Equipment        string   `json:"equipment"`
	Target           string   `json:"target"`
	SecondaryMuscles []string `json:"secondaryMuscles"`
	Instructions     []string `json:"instructions"`
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
	if cfg.S3.Bucket == "" {
		return errors.New("S3 bucket is required: set S3_BUCKET or pass --s3-bucket")
	}
	if cfg.Media.BaseURL == "" {
		return errors.New("public base URL is required: set EXERCISE_MEDIA_BASE_URL or SPORT_MAIN_DOMAIN")
	}

	tmp, err := os.MkdirTemp("", "exercises-dataset-*")
	if err != nil {
		return err
	}
	if !cfg.KeepTemp {
		defer os.RemoveAll(tmp)
	}
	repoDir := filepath.Join(tmp, "repo")
	if err := cloneDataset(ctx, cfg.RepoURL, repoDir); err != nil {
		return err
	}

	dataset, err := readDataset(repoDir)
	if err != nil {
		return err
	}
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(awsCfg)

	manifest := make([]exercises.Media, 0)
	now := time.Now().UTC().Format(time.RFC3339)
	for _, alias := range exercises.Aliases() {
		exercise, ok := resolveExercise(alias, dataset)
		if !ok {
			logx.Warn(ctx).Str("program_key", alias.ProgramKey).Str("status", string(alias.Status)).Msg("dataset match not found")
			continue
		}
		gif, err := loadGIF(ctx, repoDir, exercise)
		if err != nil {
			return fmt.Errorf("%s: load gif: %w", alias.ProgramKey, err)
		}
		key := strings.Trim(cfg.S3.Prefix, "/") + "/" + exercise.ID + ".gif"
		if _, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(cfg.S3.Bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(gif),
			ContentType: aws.String("image/gif"),
		}); err != nil {
			return fmt.Errorf("%s: upload gif: %w", alias.ProgramKey, err)
		}
		manifest = append(manifest, exercises.Media{
			DatasetID:        exercise.ID,
			GIFURL:           strings.TrimRight(cfg.Media.BaseURL, "/") + "/" + key,
			Equipment:        exercise.Equipment,
			TargetMuscles:    nonEmpty([]string{exercise.Target}),
			SecondaryMuscles: exercise.SecondaryMuscles,
			Instructions:     exercise.Instructions,
			UpdatedAt:        now,
		})
		logx.Info(ctx).Str("program_key", alias.ProgramKey).Str("dataset_id", exercise.ID).Str("key", key).Msg("uploaded exercise gif")
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
	if cfg.Media.BaseURL == "" && cfg.Media.MainDomain != "" {
		cfg.Media.BaseURL = "https://media." + cfg.Media.MainDomain
	}
	return cfg, err
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

func readDataset(repoDir string) (map[string]datasetExercise, error) {
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
		return nil, fmt.Errorf("dataset JSON not found: %w", err)
	}
	var exercisesList []datasetExercise
	if err := json.Unmarshal(data, &exercisesList); err != nil {
		return nil, err
	}
	out := make(map[string]datasetExercise, len(exercisesList))
	for _, exercise := range exercisesList {
		out[exercise.ID] = exercise
	}
	return out, nil
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

func loadGIF(ctx context.Context, repoDir string, exercise datasetExercise) ([]byte, error) {
	if u, err := url.Parse(exercise.GIFURL); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, exercise.GIFURL, nil)
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

	candidates := []string{}
	if exercise.GIFURL != "" {
		candidates = append(candidates, filepath.Join(repoDir, filepath.FromSlash(exercise.GIFURL)))
	}
	candidates = append(candidates,
		filepath.Join(repoDir, "exercises", exercise.ID+".gif"),
		filepath.Join(repoDir, "gifs", exercise.ID+".gif"),
	)
	for _, candidate := range candidates {
		if data, err := os.ReadFile(candidate); err == nil {
			return data, nil
		}
	}

	var found string
	err := filepath.WalkDir(repoDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || found != "" {
			return err
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".gif") && strings.Contains(name, exercise.ID) {
			found = p
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == "" {
		return nil, errors.New("gif file not found in dataset")
	}
	return os.ReadFile(found)
}

func writeManifest(file string, manifest []exercises.Media) error {
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
