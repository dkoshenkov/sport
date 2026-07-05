package exercises

import (
	"os"
	"path/filepath"
	"testing"

	"sport/server/internal/api"
)

func TestReverseGripBenchResolvesToAvailableMedia(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "exercise_media.json")
	manifest := `[
		{
			"datasetExerciseId": "2187",
			"gifUrl": "https://example.com/exercises/2187.gif",
			"width": 320,
			"height": 240,
			"equipment": "barbell",
			"targetMuscles": ["pectorals"]
		}
	]`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write media manifest: %v", err)
	}

	catalog, err := NewCatalog("", manifestPath)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	details, ok := catalog.Details("reverse_grip_bench")
	if !ok {
		t.Fatal("expected reverse_grip_bench alias")
	}
	if got := details.DatasetExerciseId.Or(""); got != "2187" {
		t.Fatalf("dataset id = %q, want 2187", got)
	}
	if details.Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("media status = %q, want available", details.Media.Status)
	}
	gifURL, ok := details.Media.GifUrl.Get()
	if !ok {
		t.Fatal("expected gif url")
	}
	if got := gifURL.String(); got != "https://example.com/exercises/2187.gif" {
		t.Fatalf("gif url = %q, want https://example.com/exercises/2187.gif", got)
	}
}

func TestGoodMorningUsesMediaBaseURLForReviewedAlias(t *testing.T) {
	catalog, err := NewCatalog("https://example.com", "")
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	details, ok := catalog.Details("good_morning")
	if !ok {
		t.Fatal("expected good_morning alias")
	}
	if got := details.DatasetExerciseId.Or(""); got != "0044" {
		t.Fatalf("dataset id = %q, want 0044", got)
	}
	if details.Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("media status = %q, want available", details.Media.Status)
	}
	gifURL, ok := details.Media.GifUrl.Get()
	if !ok {
		t.Fatal("expected gif url")
	}
	if got := gifURL.String(); got != "https://example.com/exercises/0044.gif" {
		t.Fatalf("gif url = %q, want https://example.com/exercises/0044.gif", got)
	}
}

func TestReviewedAliasesUseMediaBaseURL(t *testing.T) {
	catalog, err := NewCatalog("https://example.com", "")
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	tests := []struct {
		key       string
		datasetID string
	}{
		{key: "lever_vertical_row", datasetID: "0579"},
		{key: "kettlebell_military_press", datasetID: "0553"},
		{key: "one_arm_military_press", datasetID: "0361"},
		{key: "bulgarian_split_squat", datasetID: "0410"},
		{key: "low_bar_squat", datasetID: "1435"},
		{key: "high_bar_squat", datasetID: "1436"},
		{key: "zercher_squat", datasetID: "1545"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			details, ok := catalog.Details(tt.key)
			if !ok {
				t.Fatalf("expected %s alias", tt.key)
			}
			if got := details.DatasetExerciseId.Or(""); got != tt.datasetID {
				t.Fatalf("dataset id = %q, want %s", got, tt.datasetID)
			}
			if details.Media.Status != api.ExerciseMediaStatusAvailable {
				t.Fatalf("media status = %q, want available", details.Media.Status)
			}
			gifURL, ok := details.Media.GifUrl.Get()
			if !ok {
				t.Fatal("expected gif url")
			}
			want := "https://example.com/exercises/" + tt.datasetID + ".gif"
			if got := gifURL.String(); got != want {
				t.Fatalf("gif url = %q, want %s", got, want)
			}
		})
	}
}
