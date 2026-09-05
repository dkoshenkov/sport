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

func NewCatalog(datasetDir string, aliases ...Alias) (*Catalog, error) {
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
	for _, alias := range aliases {
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
