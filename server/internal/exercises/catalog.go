package exercises

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

type datasetExercise struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Category         string              `json:"category"`
	BodyPart         string              `json:"body_part"`
	Equipment        string              `json:"equipment"`
	Target           string              `json:"target"`
	SecondaryMuscles []string            `json:"secondary_muscles"`
	InstructionSteps map[string][]string `json:"instruction_steps"`
	GIFURL           string              `json:"gif_url"`
}

type Catalog struct {
	aliases      map[string]Alias
	byDatasetID  map[string]datasetExercise
	exercises    []datasetExercise
	datasetDir   string
	mediaBaseURL string
}

func NewCatalog(datasetDir string) (*Catalog, error) {
	data, err := os.ReadFile(filepath.Join(datasetDir, "data", "exercises.json"))
	if err != nil {
		return nil, fmt.Errorf("read exercise dataset: %w", err)
	}

	var list []datasetExercise
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode exercise dataset: %w", err)
	}

	c := &Catalog{
		aliases:      make(map[string]Alias),
		byDatasetID:  make(map[string]datasetExercise, len(list)),
		exercises:    make([]datasetExercise, 0, len(list)),
		datasetDir:   datasetDir,
		mediaBaseURL: "/videos",
	}
	for _, item := range list {
		if item.ID == "" {
			return nil, fmt.Errorf("exercise dataset contains an item without id")
		}
		if _, exists := c.byDatasetID[item.ID]; exists {
			return nil, fmt.Errorf("exercise dataset contains duplicate id %q", item.ID)
		}
		c.byDatasetID[item.ID] = item
		c.exercises = append(c.exercises, item)
	}
	for _, alias := range Aliases() {
		c.aliases[alias.ProgramKey] = alias
	}
	return c, nil
}

func (c *Catalog) List(params api.ListExercisesParams) api.ExerciseCatalogListResponse {
	items := make([]datasetExercise, 0, len(c.exercises))
	query := strings.ToLower(strings.TrimSpace(params.Query.Or("")))
	for _, item := range c.exercises {
		if query != "" && !c.matchesQuery(item, query) {
			continue
		}
		if params.HasImage.Or(false) && !c.hasMedia(item) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		left := strings.ToLower(items[i].Name)
		right := strings.ToLower(items[j].Name)
		if left == right {
			return items[i].ID < items[j].ID
		}
		return left < right
	})

	limit := params.Limit.Or(30)
	offset := params.Offset.Or(0)
	if limit <= 0 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	total := len(items)
	if offset >= total {
		return api.ExerciseCatalogListResponse{Items: []api.ExerciseCatalogItem{}, Total: total, Limit: limit, Offset: offset}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	result := make([]api.ExerciseCatalogItem, 0, end-offset)
	for _, item := range items[offset:end] {
		result = append(result, c.catalogItem(item))
	}
	return api.ExerciseCatalogListResponse{Items: result, Total: total, Limit: limit, Offset: offset}
}

func (c *Catalog) CatalogExercise(datasetID string) (api.ExerciseCatalogItem, bool) {
	item, ok := c.byDatasetID[datasetID]
	if !ok {
		return api.ExerciseCatalogItem{}, false
	}
	return c.catalogItem(item), true
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
		Media:            api.ExerciseMedia{Status: api.ExerciseMediaStatusMissing},
	}
	item, ok := c.resolveAlias(alias)
	if !ok {
		return details, true
	}
	details.DatasetExerciseId = api.NewOptNilString(item.ID)
	details.DatasetName = api.NewOptNilString(item.Name)
	details.Equipment = api.NewOptNilString(item.Equipment)
	details.TargetMuscles = nonEmpty([]string{item.Target})
	details.SecondaryMuscles = append([]string(nil), item.SecondaryMuscles...)
	details.Instructions = preferredInstructions(item)
	details.Media = c.media(item)
	return details, true
}

func (c *Catalog) matchesQuery(item datasetExercise, query string) bool {
	values := []string{item.ID, item.Name, item.Category, item.BodyPart, item.Equipment, item.Target}
	values = append(values, item.SecondaryMuscles...)
	for _, alias := range c.aliases {
		resolved, ok := c.resolveAlias(alias)
		if ok && resolved.ID == item.ID {
			values = append(values, alias.ProgramKey, alias.ProgramName)
			values = append(values, alias.NameHints...)
		}
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

func (c *Catalog) catalogItem(item datasetExercise) api.ExerciseCatalogItem {
	result := api.ExerciseCatalogItem{
		DatasetExerciseId: item.ID,
		Name:              item.Name,
		Category:          api.NewOptNilString(item.Category),
		BodyPart:          api.NewOptNilString(item.BodyPart),
		Equipment:         api.NewOptNilString(item.Equipment),
		TargetMuscles:     nonEmpty([]string{item.Target}),
		SecondaryMuscles:  append([]string(nil), item.SecondaryMuscles...),
		Instructions:      preferredInstructions(item),
		Media:             c.media(item),
	}
	for _, alias := range c.aliases {
		resolved, ok := c.resolveAlias(alias)
		if ok && resolved.ID == item.ID {
			result.NameRu = api.NewOptNilString(alias.ProgramName)
			break
		}
	}
	return result
}

func (c *Catalog) media(item datasetExercise) api.ExerciseMedia {
	media := api.ExerciseMedia{Status: api.ExerciseMediaStatusMissing}
	if !c.hasMedia(item) {
		return media
	}
	imageURL, err := url.Parse(c.mediaBaseURL + "/" + filepath.Base(item.GIFURL))
	if err != nil {
		return media
	}
	media.Status = api.ExerciseMediaStatusAvailable
	media.ImageUrl = api.NewOptNilURI(*imageURL)
	media.Provenance = api.NewOptNilString("exercises-dataset-main")
	return media
}

func (c *Catalog) hasMedia(item datasetExercise) bool {
	if item.GIFURL == "" {
		return false
	}
	path := filepath.Clean(item.GIFURL)
	if filepath.IsAbs(path) || path == "." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
		return false
	}
	info, err := os.Stat(filepath.Join(c.datasetDir, path))
	return err == nil && !info.IsDir()
}

func (c *Catalog) resolveAlias(alias Alias) (datasetExercise, bool) {
	if alias.DatasetID != "" {
		item, ok := c.byDatasetID[alias.DatasetID]
		return item, ok
	}
	if alias.Status == StatusMissing {
		return datasetExercise{}, false
	}
	hints := normalizedSet(alias.NameHints)
	for _, item := range c.exercises {
		if _, ok := hints[normalize(item.Name)]; ok {
			return item, true
		}
	}
	return datasetExercise{}, false
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
	return []string{}
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func normalizedSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[normalize(value)] = struct{}{}
	}
	return result
}

func normalize(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(value, "-", " "))), " ")
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
		{ProgramKey: "paused_deadlift", ProgramName: "Становая тяга с паузами", Status: StatusMissing},

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

		{ProgramKey: "abs", ProgramName: "Упражнение на пресс", Status: StatusMissing},
		{ProgramKey: "triceps", ProgramName: "Трицепс", Status: StatusMissing},
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
