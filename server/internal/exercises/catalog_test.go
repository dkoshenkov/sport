package exercises

import (
	"path/filepath"
	"testing"

	"sport/server/internal/api"
)

func TestCatalogReadsExercisesDatasetDirectly(t *testing.T) {
	catalog, err := NewCatalog(filepath.Join("..", "..", "..", "exercises-dataset-main"))
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	response := catalog.List(api.ListExercisesParams{
		Query: api.NewOptString("barbell deadlift"),
	})
	if response.Total == 0 {
		t.Fatal("dataset query returned no exercises")
	}
	if response.Items[0].DatasetExerciseId != "0032" {
		t.Fatalf("first item id = %q, want 0032", response.Items[0].DatasetExerciseId)
	}
	if response.Items[0].Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("media status = %q, want available", response.Items[0].Media.Status)
	}
	imageURL, ok := response.Items[0].Media.ImageUrl.Get()
	if !ok || imageURL.String() != "/videos/0032-ila4NZS.gif" {
		t.Fatalf("image URL = %q, %v; want /videos/0032-ila4NZS.gif", imageURL, ok)
	}
}

func TestCatalogResolvesProgramAliasFromDataset(t *testing.T) {
	catalog, err := NewCatalog(filepath.Join("..", "..", "..", "exercises-dataset-main"))
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	details, ok := catalog.Details("deadlift")
	if !ok {
		t.Fatal("expected deadlift alias")
	}
	if got := details.DatasetExerciseId.Or(""); got != "0032" {
		t.Fatalf("dataset id = %q, want 0032", got)
	}
	if got := details.DatasetName.Or(""); got != "barbell deadlift" {
		t.Fatalf("dataset name = %q, want barbell deadlift", got)
	}
	if details.Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("media status = %q, want available", details.Media.Status)
	}
	if len(details.Instructions) == 0 {
		t.Fatal("expected instructions from dataset")
	}
}

func TestCatalogMediaFilterUsesDatasetGIFs(t *testing.T) {
	catalog, err := NewCatalog(filepath.Join("..", "..", "..", "exercises-dataset-main"))
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	response := catalog.List(api.ListExercisesParams{HasImage: api.NewOptBool(true), Limit: api.NewOptInt(100)})
	if response.Total != 1324 {
		t.Fatalf("GIF-backed exercises = %d, want 1324", response.Total)
	}
}
