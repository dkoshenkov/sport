package exercises

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"sport/server/internal/api"
)

type AliasStatus string

const (
	StatusConfirmed   AliasStatus = "confirmed"
	StatusMissing     AliasStatus = "missing"
	StatusNeedsReview AliasStatus = "needs_review"
)

type Alias struct {
	ProgramKey  string      `json:"programExerciseKey"`
	ProgramName string      `json:"programNameRu"`
	DatasetID   string      `json:"datasetExerciseId,omitempty"`
	DatasetName string      `json:"datasetName,omitempty"`
	Status      AliasStatus `json:"reviewStatus"`
	Notes       string      `json:"notes,omitempty"`
	NameHints   []string    `json:"nameHints,omitempty"`
}

type Media struct {
	DatasetID        string   `json:"datasetExerciseId"`
	GIFURL           string   `json:"gifUrl"`
	Width            int      `json:"width,omitempty"`
	Height           int      `json:"height,omitempty"`
	Equipment        string   `json:"equipment,omitempty"`
	TargetMuscles    []string `json:"targetMuscles,omitempty"`
	SecondaryMuscles []string `json:"secondaryMuscles,omitempty"`
	Instructions     []string `json:"instructions,omitempty"`
	UpdatedAt        string   `json:"updatedAt,omitempty"`
}

type Catalog struct {
	aliases     map[string]Alias
	media       map[string]Media
	mediaBase   string
	mediaPrefix string
}

func NewCatalog(mediaBaseURL, mediaManifestPath string) (*Catalog, error) {
	c := &Catalog{
		aliases:     map[string]Alias{},
		media:       map[string]Media{},
		mediaBase:   strings.TrimRight(mediaBaseURL, "/"),
		mediaPrefix: "exercises",
	}
	for _, alias := range Aliases() {
		c.aliases[alias.ProgramKey] = alias
	}
	if mediaManifestPath != "" {
		if err := c.loadMediaManifest(mediaManifestPath); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Catalog) Details(exerciseKey string) (api.ExerciseDetails, bool) {
	alias, ok := c.aliases[exerciseKey]
	if !ok {
		return api.ExerciseDetails{}, false
	}

	details := api.ExerciseDetails{
		ExerciseKey:      alias.ProgramKey,
		Name:             alias.ProgramName,
		AliasStatus:      api.ExerciseDetailsAliasStatus(alias.Status),
		TargetMuscles:    []string{},
		SecondaryMuscles: []string{},
		Instructions:     []string{},
		Media: api.ExerciseMedia{
			Status: api.ExerciseMediaStatusMissing,
		},
	}
	if alias.DatasetID != "" {
		details.DatasetExerciseId = api.NewOptNilString(alias.DatasetID)
	}

	if media, ok := c.media[alias.DatasetID]; ok {
		applyMedia(&details, media)
		return details, true
	}
	if alias.Status == StatusConfirmed && alias.DatasetID != "" && c.mediaBase != "" {
		gifURL, err := url.Parse(c.mediaBase + "/" + path.Join(c.mediaPrefix, alias.DatasetID+".gif"))
		if err == nil {
			details.Media = api.ExerciseMedia{
				Status: api.ExerciseMediaStatusAvailable,
				GifUrl: api.NewOptNilURI(*gifURL),
			}
		}
	}
	return details, true
}

func (c *Catalog) loadMediaManifest(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read exercise media manifest: %w", err)
	}
	var items []Media
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("decode exercise media manifest: %w", err)
	}
	for _, item := range items {
		if item.DatasetID == "" {
			return errors.New("exercise media manifest contains item without datasetExerciseId")
		}
		c.media[item.DatasetID] = item
	}
	return nil
}

func applyMedia(details *api.ExerciseDetails, media Media) {
	if media.Equipment != "" {
		details.Equipment = api.NewOptNilString(media.Equipment)
	}
	details.TargetMuscles = append([]string(nil), media.TargetMuscles...)
	details.SecondaryMuscles = append([]string(nil), media.SecondaryMuscles...)
	details.Instructions = append([]string(nil), media.Instructions...)
	if media.GIFURL == "" {
		return
	}
	gifURL, err := url.Parse(media.GIFURL)
	if err != nil {
		return
	}
	details.Media = api.ExerciseMedia{
		Status: api.ExerciseMediaStatusAvailable,
		GifUrl: api.NewOptNilURI(*gifURL),
	}
	if media.Width > 0 {
		details.Media.Width = api.NewOptNilInt(media.Width)
	}
	if media.Height > 0 {
		details.Media.Height = api.NewOptNilInt(media.Height)
	}
}

func Aliases() []Alias {
	items := []Alias{
		{ProgramKey: "deadlift", ProgramName: "Становая тяга", DatasetID: "0032", DatasetName: "barbell deadlift", Status: StatusConfirmed, NameHints: []string{"barbell deadlift"}},
		{ProgramKey: "bench_press", ProgramName: "Жим лежа", DatasetID: "0025", DatasetName: "barbell bench press", Status: StatusConfirmed, NameHints: []string{"barbell bench press"}},
		{ProgramKey: "squat", ProgramName: "Приседания", DatasetID: "0043", DatasetName: "barbell full squat", Status: StatusConfirmed, NameHints: []string{"barbell full squat"}},
		{ProgramKey: "pull_up", ProgramName: "Подтягивания", DatasetID: "0652", DatasetName: "pull-up", Status: StatusConfirmed, NameHints: []string{"pull-up"}},

		{ProgramKey: "good_morning", ProgramName: "Гуд-морнинг", Status: StatusNeedsReview, NameHints: []string{"good morning"}},
		{ProgramKey: "romanian_deadlift", ProgramName: "Румынская тяга", Status: StatusNeedsReview, NameHints: []string{"romanian deadlift"}},
		{ProgramKey: "deficit_deadlift", ProgramName: "Тяга из ямы", Status: StatusNeedsReview, NameHints: []string{"deficit deadlift"}},
		{ProgramKey: "sumo_deadlift", ProgramName: "Становая тяга сумо", Status: StatusNeedsReview, NameHints: []string{"sumo deadlift"}},
		{ProgramKey: "paused_deadlift", ProgramName: "Становая тяга с паузами", Status: StatusMissing, Notes: "No confirmed exact dataset match."},

		{ProgramKey: "close_grip_bench_press", ProgramName: "Жим узким хватом", Status: StatusNeedsReview, NameHints: []string{"close-grip bench press", "close grip bench press"}},
		{ProgramKey: "reverse_grip_bench_press", ProgramName: "Жим обратным хватом", Status: StatusNeedsReview, NameHints: []string{"reverse grip bench press"}},
		{ProgramKey: "incline_bench_press", ProgramName: "Жим на наклонной скамье", Status: StatusNeedsReview, NameHints: []string{"incline bench press"}},
		{ProgramKey: "dumbbell_bench_press", ProgramName: "Жим гантелей лежа", Status: StatusNeedsReview, NameHints: []string{"dumbbell bench press"}},
		{ProgramKey: "dip", ProgramName: "Брусья", Status: StatusNeedsReview, NameHints: []string{"chest dip", "dip"}},

		{ProgramKey: "zercher_squat", ProgramName: "Приседания Зерчера", Status: StatusNeedsReview, NameHints: []string{"zercher squat"}},
		{ProgramKey: "front_squat", ProgramName: "Приседания со штангой на груди", Status: StatusNeedsReview, NameHints: []string{"front squat"}},
		{ProgramKey: "high_bar_squat", ProgramName: "Приседания с высоким грифом", Status: StatusMissing, Notes: "Dataset has generic squat variants; high-bar needs manual review."},
		{ProgramKey: "low_bar_squat", ProgramName: "Приседания с низким грифом", Status: StatusMissing, Notes: "Dataset has generic squat variants; low-bar needs manual review."},
		{ProgramKey: "bulgarian_split_squat", ProgramName: "Болгарские сплит-приседания", Status: StatusNeedsReview, NameHints: []string{"bulgarian split squat"}},

		{ProgramKey: "abs", ProgramName: "Упражнение на пресс", Status: StatusMissing, Notes: "Generic category, not a single dataset exercise."},
		{ProgramKey: "triceps", ProgramName: "Трицепс", Status: StatusMissing, Notes: "Generic category, not a single dataset exercise."},
		{ProgramKey: "biceps", ProgramName: "Бицепс", DatasetID: "0294", DatasetName: "barbell curl", Status: StatusConfirmed, NameHints: []string{"barbell curl"}},
		{ProgramKey: "bent_over_barbell_row", ProgramName: "Тяга штанги в наклоне", Status: StatusNeedsReview, NameHints: []string{"bent over barbell row", "barbell bent over row"}},
		{ProgramKey: "seated_cable_row", ProgramName: "Горизонтальный блок", Status: StatusNeedsReview, NameHints: []string{"seated cable row", "cable seated row"}},
		{ProgramKey: "one_arm_dumbbell_row", ProgramName: "Тяга гантели в наклоне", Status: StatusNeedsReview, NameHints: []string{"one arm dumbbell row", "dumbbell row"}},
		{ProgramKey: "lever_horizontal_row", ProgramName: "Рычажная горизонтальная тяга", Status: StatusNeedsReview, NameHints: []string{"lever seated row", "lever row"}},
		{ProgramKey: "lat_pulldown", ProgramName: "Вертикальный блок", Status: StatusNeedsReview, NameHints: []string{"cable pulldown", "lat pulldown"}},
		{ProgramKey: "lever_vertical_row", ProgramName: "Рычажная вертикальная тяга", Status: StatusNeedsReview, NameHints: []string{"lever pulldown", "lever vertical row"}},
		{ProgramKey: "dumbbell_overhead_press", ProgramName: "Армейский жим гантелей", Status: StatusNeedsReview, NameHints: []string{"dumbbell shoulder press", "dumbbell overhead press"}},
		{ProgramKey: "handstand_push_up", ProgramName: "Отжимания в стойке на руках", Status: StatusNeedsReview, NameHints: []string{"handstand push-up"}},
		{ProgramKey: "kettlebell_overhead_press", ProgramName: "Армейский жим гирь", Status: StatusNeedsReview, NameHints: []string{"kettlebell clean and press", "kettlebell press"}},
		{ProgramKey: "one_arm_overhead_press", ProgramName: "Армейский жим одной рукой", Status: StatusNeedsReview, NameHints: []string{"one arm dumbbell press", "single arm shoulder press"}},
		{ProgramKey: "barbell_overhead_press", ProgramName: "Армейский жим штанги", Status: StatusNeedsReview, NameHints: []string{"barbell standing military press", "barbell shoulder press"}},
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ProgramKey < items[j].ProgramKey })
	return items
}
