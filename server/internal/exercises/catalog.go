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
	StorageKey       string   `json:"storageKey,omitempty"`
	Width            int      `json:"width,omitempty"`
	Height           int      `json:"height,omitempty"`
	Provenance       string   `json:"provenance,omitempty"`
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
	if alias.DatasetName != "" {
		details.DatasetName = api.NewOptNilString(alias.DatasetName)
	}

	if media, ok := c.media[alias.DatasetID]; ok {
		applyMedia(&details, media)
		return details, true
	}
	if alias.DatasetID != "" && c.mediaBase != "" {
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

func (c *Catalog) DatasetIDsForQuery(query string, limit int) []string {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return nil
	}
	if limit <= 0 {
		limit = 10
	}

	seen := map[string]bool{}
	ids := []string{}
	for _, alias := range Aliases() {
		if alias.DatasetID == "" || seen[alias.DatasetID] {
			continue
		}
		if aliasMatchesQuery(alias, needle) {
			seen[alias.DatasetID] = true
			ids = append(ids, alias.DatasetID)
			if len(ids) >= limit {
				return ids
			}
		}
	}
	return ids
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

func aliasMatchesQuery(alias Alias, needle string) bool {
	values := []string{
		alias.ProgramKey,
		alias.ProgramName,
		alias.DatasetID,
		alias.DatasetName,
	}
	values = append(values, alias.NameHints...)
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
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
	if media.StorageKey != "" {
		details.Media.StorageKey = api.NewOptNilString(media.StorageKey)
	}
	if media.Provenance != "" {
		details.Media.Provenance = api.NewOptNilString(media.Provenance)
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

		{ProgramKey: "good_morning", ProgramName: "Гуд-морнинг", DatasetID: "0044", DatasetName: "barbell good morning", Status: StatusNeedsReview, NameHints: []string{"barbell good morning", "good morning"}},
		{ProgramKey: "romanian_deadlift", ProgramName: "Румынская тяга", DatasetID: "0085", DatasetName: "barbell romanian deadlift", Status: StatusNeedsReview, NameHints: []string{"barbell romanian deadlift", "romanian deadlift"}},
		{ProgramKey: "deficit_deadlift", ProgramName: "Тяга из ямы", Status: StatusNeedsReview, NameHints: []string{"deficit deadlift"}},
		{ProgramKey: "sumo_deadlift", ProgramName: "Становая тяга сумо", Status: StatusNeedsReview, NameHints: []string{"sumo deadlift"}},
		{ProgramKey: "paused_deadlift", ProgramName: "Становая тяга с паузами", Status: StatusMissing, Notes: "No confirmed exact dataset match."},

		{ProgramKey: "close_grip_bench", ProgramName: "Жим узким хватом", DatasetID: "0030", DatasetName: "barbell close-grip bench press", Status: StatusNeedsReview, NameHints: []string{"barbell close-grip bench press", "close-grip bench press", "close grip bench press"}},
		{ProgramKey: "reverse_grip_bench", ProgramName: "Жим обратным хватом", DatasetID: "2187", DatasetName: "barbell reverse close-grip bench press", Status: StatusNeedsReview, NameHints: []string{"barbell reverse close-grip bench press", "reverse grip bench press"}},
		{ProgramKey: "incline_bench", ProgramName: "Жим на наклонной скамье", DatasetID: "0047", DatasetName: "barbell incline bench press", Status: StatusNeedsReview, NameHints: []string{"barbell incline bench press", "incline bench press"}},
		{ProgramKey: "dumbbell_bench", ProgramName: "Жим гантелей лежа", DatasetID: "0289", DatasetName: "dumbbell bench press", Status: StatusNeedsReview, NameHints: []string{"dumbbell bench press"}},
		{ProgramKey: "dips", ProgramName: "Брусья", DatasetID: "0251", DatasetName: "chest dip", Status: StatusNeedsReview, NameHints: []string{"chest dip", "dip"}},

		{ProgramKey: "zercher_squat", ProgramName: "Приседания Зерчера", DatasetID: "1545", DatasetName: "barbell full zercher squat", Status: StatusNeedsReview, NameHints: []string{"barbell full zercher squat", "zercher squat"}},
		{ProgramKey: "front_squat", ProgramName: "Приседания со штангой на груди", DatasetID: "0042", DatasetName: "barbell front squat", Status: StatusNeedsReview, NameHints: []string{"barbell front squat", "front squat"}},
		{ProgramKey: "high_bar_squat", ProgramName: "Приседания с высоким грифом", DatasetID: "1436", DatasetName: "barbell high bar squat", Status: StatusNeedsReview, NameHints: []string{"barbell high bar squat", "high bar squat"}},
		{ProgramKey: "low_bar_squat", ProgramName: "Приседания с низким грифом", DatasetID: "1435", DatasetName: "barbell low bar squat", Status: StatusNeedsReview, NameHints: []string{"barbell low bar squat", "low bar squat"}},
		{ProgramKey: "bulgarian_split_squat", ProgramName: "Болгарские сплит-приседания", DatasetID: "0410", DatasetName: "dumbbell single leg split squat", Status: StatusNeedsReview, NameHints: []string{"dumbbell single leg split squat", "bulgarian split squat", "rear foot elevated split squat"}},

		{ProgramKey: "abs", ProgramName: "Упражнение на пресс", Status: StatusMissing, Notes: "Generic category, not a single dataset exercise."},
		{ProgramKey: "triceps", ProgramName: "Трицепс", Status: StatusMissing, Notes: "Generic category, not a single dataset exercise."},
		{ProgramKey: "biceps", ProgramName: "Бицепс", DatasetID: "0294", DatasetName: "barbell curl", Status: StatusConfirmed, NameHints: []string{"barbell curl"}},
		{ProgramKey: "barbell_row", ProgramName: "Тяга штанги в наклоне", DatasetID: "0027", DatasetName: "barbell bent over row", Status: StatusNeedsReview, NameHints: []string{"barbell bent over row", "bent over barbell row"}},
		{ProgramKey: "cable_seated_row", ProgramName: "Горизонтальный блок", Status: StatusNeedsReview, NameHints: []string{"seated cable row", "cable seated row"}},
		{ProgramKey: "dumbbell_row", ProgramName: "Тяга гантели в наклоне", Status: StatusNeedsReview, NameHints: []string{"one arm dumbbell row", "dumbbell row"}},
		{ProgramKey: "lever_horizontal_row", ProgramName: "Рычажная горизонтальная тяга", Status: StatusNeedsReview, NameHints: []string{"lever seated row", "lever row"}},
		{ProgramKey: "lat_pulldown", ProgramName: "Вертикальный блок", DatasetID: "0150", DatasetName: "cable bar lateral pulldown", Status: StatusNeedsReview, NameHints: []string{"cable bar lateral pulldown", "cable pulldown", "lat pulldown"}},
		{ProgramKey: "lever_vertical_row", ProgramName: "Рычажная вертикальная тяга", DatasetID: "0579", DatasetName: "lever front pulldown", Status: StatusNeedsReview, NameHints: []string{"lever front pulldown", "lever pulldown", "lever vertical row"}},
		{ProgramKey: "dumbbell_military_press", ProgramName: "Армейский жим гантелей", DatasetID: "0405", DatasetName: "dumbbell seated shoulder press", Status: StatusNeedsReview, NameHints: []string{"dumbbell seated shoulder press", "dumbbell shoulder press", "dumbbell overhead press"}},
		{ProgramKey: "handstand_push_up", ProgramName: "Отжимания в стойке на руках", Status: StatusNeedsReview, NameHints: []string{"handstand push-up"}},
		{ProgramKey: "kettlebell_military_press", ProgramName: "Армейский жим гирь", DatasetID: "0553", DatasetName: "kettlebell two arm military press", Status: StatusNeedsReview, NameHints: []string{"kettlebell two arm military press", "kettlebell clean and press", "kettlebell press"}},
		{ProgramKey: "one_arm_military_press", ProgramName: "Армейский жим одной рукой", DatasetID: "0361", DatasetName: "dumbbell one arm shoulder press", Status: StatusNeedsReview, NameHints: []string{"dumbbell one arm shoulder press", "one arm dumbbell press", "single arm shoulder press"}},
		{ProgramKey: "barbell_military_press", ProgramName: "Армейский жим штанги", DatasetID: "1456", DatasetName: "barbell standing close grip military press", Status: StatusNeedsReview, NameHints: []string{"barbell standing close grip military press", "barbell standing military press", "barbell shoulder press"}},
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ProgramKey < items[j].ProgramKey })
	return items
}
