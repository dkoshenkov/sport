package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"sport/server/internal/exercises"
)

func TestPrepareDatasetDirUsesLocalDir(t *testing.T) {
	dir := t.TempDir()
	got, err := prepareDatasetDir(t.Context(), syncConfig{DatasetDir: dir}, t.TempDir())
	if err != nil {
		t.Fatalf("prepareDatasetDir() error = %v", err)
	}
	if got != dir {
		t.Fatalf("prepareDatasetDir() = %q, want %q", got, dir)
	}
}

func TestReadDatasetCurrentShape(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "exercises.json"), []byte(`[
		{
			"id": "0025",
			"name": "barbell bench press",
			"category": "chest",
			"body_part": "chest",
			"equipment": "barbell",
			"instructions": {"en": "Press.", "tr": "Bastir."},
			"instruction_steps": {"en": ["Set up.", "Press."], "tr": ["Kur.", "Bastir."]},
			"muscle_group": "triceps",
			"secondary_muscles": ["triceps", "shoulders"],
			"target": "pectorals",
			"image": "images/0025-EIeI8Vf.jpg",
			"gif_url": "videos/0025-EIeI8Vf.gif",
			"created_at": "2026-03-18 12:31:32+00:00"
		}
	]`), 0o644); err != nil {
		t.Fatal(err)
	}

	list, byID, err := readDataset(dir)
	if err != nil {
		t.Fatalf("readDataset() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
	exercise := byID["0025"]
	if exercise.GIFURL != "videos/0025-EIeI8Vf.gif" {
		t.Fatalf("gif url = %q", exercise.GIFURL)
	}
	if got := preferredInstructions(exercise); len(got) != 2 || got[0] != "Set up." {
		t.Fatalf("preferred instructions = %#v", got)
	}
}

func TestResolveExerciseByDatasetIDAndHints(t *testing.T) {
	dataset := map[string]datasetExercise{
		"0025": {ID: "0025", Name: "barbell bench press"},
		"9999": {ID: "9999", Name: "close grip bench press"},
	}

	byID, ok := resolveExercise(exercises.Alias{DatasetID: "0025"}, dataset)
	if !ok || byID.ID != "0025" {
		t.Fatalf("resolve by id = %#v, %v", byID, ok)
	}

	byHint, ok := resolveExercise(exercises.Alias{
		Status:    exercises.StatusNeedsReview,
		NameHints: []string{"close-grip bench press"},
	}, dataset)
	if !ok || byHint.ID != "9999" {
		t.Fatalf("resolve by hint = %#v, %v", byHint, ok)
	}
}

func TestLoadGIFMissingDoesNotFail(t *testing.T) {
	_, ok, err := loadGIF(t.Context(), syncConfig{}, t.TempDir(), datasetExercise{
		ID:     "0001",
		GIFURL: "videos/0001-missing.gif",
	})
	if err != nil {
		t.Fatalf("loadGIF() error = %v", err)
	}
	if ok {
		t.Fatal("loadGIF() ok = true, want false")
	}
}

func TestLoadGIFFallsBackToMediaIDURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2gPfomN.gif" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/gif")
		_, _ = w.Write([]byte("gif"))
	}))
	t.Cleanup(server.Close)

	data, ok, err := loadGIF(t.Context(), syncConfig{
		Storage: syncStorageConfig{SourceBaseURL: server.URL},
	}, t.TempDir(), datasetExercise{
		ID:     "0001",
		GIFURL: "videos/0001-2gPfomN.gif",
	})
	if err != nil {
		t.Fatalf("loadGIF() error = %v", err)
	}
	if !ok {
		t.Fatal("loadGIF() ok = false, want true")
	}
	if string(data) != "gif" {
		t.Fatalf("loadGIF() data = %q", data)
	}
}

func TestLoadGIFUsesMediaIDWhenGIFURLIsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2gPfomN.gif" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/gif")
		_, _ = w.Write([]byte("gif-by-media-id"))
	}))
	t.Cleanup(server.Close)

	data, ok, err := loadGIF(t.Context(), syncConfig{
		Storage: syncStorageConfig{SourceBaseURL: server.URL},
	}, t.TempDir(), datasetExercise{
		ID:      "0001",
		MediaID: "2gPfomN",
	})
	if err != nil {
		t.Fatalf("loadGIF() error = %v", err)
	}
	if !ok {
		t.Fatal("loadGIF() ok = false, want true")
	}
	if string(data) != "gif-by-media-id" {
		t.Fatalf("loadGIF() data = %q", data)
	}
}

func TestLoadGIFUsesLocalSourceDirByDatasetID(t *testing.T) {
	sourceDir := t.TempDir()
	assetsDir := filepath.Join(sourceDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "0001.gif"), []byte("local-gif"), 0o644); err != nil {
		t.Fatal(err)
	}

	data, ok, err := loadGIF(t.Context(), syncConfig{
		Storage: syncStorageConfig{SourceDir: sourceDir},
	}, t.TempDir(), datasetExercise{
		ID:      "0001",
		MediaID: "2gPfomN",
	})
	if err != nil {
		t.Fatalf("loadGIF() error = %v", err)
	}
	if !ok {
		t.Fatal("loadGIF() ok = false, want true")
	}
	if string(data) != "local-gif" {
		t.Fatalf("loadGIF() data = %q", data)
	}
}

func TestLocalGIFCandidates(t *testing.T) {
	got := localGIFCandidates("/gifs", datasetExercise{
		ID:      "0001",
		MediaID: "2gPfomN",
		GIFURL:  "videos/0001-old.gif",
	})
	want := []string{
		"/gifs/0001.gif",
		"/gifs/assets/0001.gif",
		"/gifs/2gPfomN.gif",
		"/gifs/assets/2gPfomN.gif",
		"/gifs/0001-old.gif",
		"/gifs/assets/0001-old.gif",
	}
	if len(got) != len(want) {
		t.Fatalf("len(localGIFCandidates) = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidate[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWriteManifestCreatesParentDir(t *testing.T) {
	file := filepath.Join(t.TempDir(), "nested", "exercise-media.json")
	if err := writeManifest(file, []exercises.Media{{DatasetID: "0001"}}); err != nil {
		t.Fatalf("writeManifest() error = %v", err)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("manifest stat error = %v", err)
	}
}

func TestPublicMediaURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		key     string
		want    string
	}{
		{name: "relative", key: "exercises/0032.gif", want: "/exercises/0032.gif"},
		{name: "absolute", baseURL: "http://127.0.0.1:38174", key: "exercises/0032.gif", want: "http://127.0.0.1:38174/exercises/0032.gif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publicMediaURL(tt.baseURL, tt.key); got != tt.want {
				t.Fatalf("publicMediaURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMediaIDFromGIFPath(t *testing.T) {
	tests := map[string]string{
		"videos/0001-2gPfomN.gif": "2gPfomN",
		"2gPfomN.gif":             "2gPfomN",
		"":                        "",
	}
	for input, want := range tests {
		if got := mediaIDFromGIFPath(input); got != want {
			t.Fatalf("mediaIDFromGIFPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestMediaIDFromExercise(t *testing.T) {
	tests := []struct {
		name     string
		exercise datasetExercise
		want     string
	}{
		{
			name:     "media id",
			exercise: datasetExercise{MediaID: "2gPfomN", GIFURL: "videos/0001-other.gif"},
			want:     "2gPfomN",
		},
		{
			name:     "gif path",
			exercise: datasetExercise{GIFURL: "videos/0001-2gPfomN.gif"},
			want:     "2gPfomN",
		},
		{
			name:     "image path",
			exercise: datasetExercise{Image: "images/0001-2gPfomN.jpg"},
			want:     "2gPfomN",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mediaIDFromExercise(tt.exercise); got != tt.want {
				t.Fatalf("mediaIDFromExercise() = %q, want %q", got, tt.want)
			}
		})
	}
}
