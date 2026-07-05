package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"sport/server/internal/api"
	"sport/server/internal/exercises"
)

func TestRegisterHashesPasswordAndSetsSessionCookie(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(store, nil)

	res, err := handler.Register(context.Background(), &api.RegisterRequest{
		Nickname: "athlete_1",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	auth, ok := res.(*api.AuthResponseHeaders)
	if !ok {
		t.Fatalf("Register() response = %T, want *api.AuthResponseHeaders", res)
	}
	cookie, ok := auth.SetCookie.Get()
	if !ok || !strings.Contains(cookie, "sid=") || !strings.Contains(cookie, "HttpOnly") {
		t.Fatalf("SetCookie = %q, ok = %v; want sid HttpOnly cookie", cookie, ok)
	}
	if store.passwordHash == "password123" {
		t.Fatal("stored raw password")
	}
	if bcrypt.CompareHashAndPassword([]byte(store.passwordHash), []byte("password123")) != nil {
		t.Fatal("stored password hash does not match password")
	}
	if store.sessionTokenHash == "" {
		t.Fatal("session token hash was not stored")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	store := newFakeStore()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	store.user = api.User{ID: uuid.New(), Nickname: "athlete_1", CreatedAt: time.Now()}
	store.passwordHash = string(hash)
	handler := NewHandler(store, nil)

	res, err := handler.Login(context.Background(), &api.LoginRequest{
		Nickname: "athlete_1",
		Password: "wrong-password",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if _, ok := res.(*api.LoginUnauthorized); !ok {
		t.Fatalf("Login() response = %T, want *api.LoginUnauthorized", res)
	}
}

func TestGetAuthSessionReturnsUnauthenticatedWithoutCookie(t *testing.T) {
	store := newFakeStore()
	handler := NewHandler(store, nil)

	session, err := handler.GetAuthSession(context.Background(), api.GetAuthSessionParams{})
	if err != nil {
		t.Fatalf("GetAuthSession() error = %v", err)
	}
	if session.Authenticated {
		t.Fatal("Authenticated = true, want false")
	}
}

func TestGetAuthSessionReturnsUserForValidCookie(t *testing.T) {
	store := newFakeStore()
	store.user = api.User{ID: uuid.New(), Nickname: "athlete_1", CreatedAt: time.Now()}
	token := "session-token"
	store.sessions = map[string]uuid.UUID{hashToken(token): store.user.ID}
	handler := NewHandler(store, nil)

	session, err := handler.GetAuthSession(context.Background(), api.GetAuthSessionParams{Sid: api.NewOptString(token)})
	if err != nil {
		t.Fatalf("GetAuthSession() error = %v", err)
	}
	if !session.Authenticated {
		t.Fatal("Authenticated = false, want true")
	}
	user, ok := session.User.Get()
	if !ok || user.ID != store.user.ID {
		t.Fatalf("session user = %#v, %v; want %s", user, ok, store.user.ID)
	}
}

func TestListCyclesEmpty(t *testing.T) {
	store := newFakeStore()
	store.user = api.User{ID: uuid.New(), Nickname: "athlete_1", CreatedAt: time.Now()}
	handler := NewHandler(store, nil)

	res, err := handler.ListCycles(authContext(store.user.ID))
	if err != nil {
		t.Fatalf("ListCycles() error = %v", err)
	}
	cycles, ok := res.(*api.CyclesResponse)
	if !ok {
		t.Fatalf("ListCycles() response = %T, want *api.CyclesResponse", res)
	}
	if len(cycles.Cycles) != 0 {
		t.Fatalf("cycles = %#v, want empty", cycles.Cycles)
	}
	if _, ok := cycles.CurrentCycleId.Get(); ok {
		t.Fatal("currentCycleId is set, want null")
	}
}

func TestCreateCycleArchivesPreviousActive(t *testing.T) {
	store := newFakeStore()
	store.user = api.User{ID: uuid.New(), Nickname: "athlete_1", CreatedAt: time.Now()}
	oldID := uuid.New()
	store.cycles = []api.ProgramCycle{{ID: oldID, Title: "Old", Status: api.CycleStatusActive, CurrentWeek: api.ProgramWeekWeek1, Settings: testCycleSettings()}}
	handler := NewHandler(store, nil)

	res, err := handler.CreateCycle(authContext(store.user.ID), &api.PutCurrentCycleRequest{
		Title:       "New",
		CurrentWeek: api.ProgramWeekWeek1,
		Settings:    testCycleSettings(),
	})
	if err != nil {
		t.Fatalf("CreateCycle() error = %v", err)
	}
	created, ok := res.(*api.CycleResponse)
	if !ok {
		t.Fatalf("CreateCycle() response = %T, want *api.CycleResponse", res)
	}
	if created.Cycle.Status != api.CycleStatusActive {
		t.Fatalf("new status = %s, want active", created.Cycle.Status)
	}
	if store.cycles[1].Status != api.CycleStatusArchived {
		t.Fatalf("old status = %s, want archived", store.cycles[1].Status)
	}
}

func TestActivateCycleArchivesPreviousActive(t *testing.T) {
	store := newFakeStore()
	store.user = api.User{ID: uuid.New(), Nickname: "athlete_1", CreatedAt: time.Now()}
	firstID := uuid.New()
	secondID := uuid.New()
	store.cycles = []api.ProgramCycle{
		{ID: firstID, Title: "First", Status: api.CycleStatusActive, CurrentWeek: api.ProgramWeekWeek1, Settings: testCycleSettings()},
		{ID: secondID, Title: "Second", Status: api.CycleStatusArchived, CurrentWeek: api.ProgramWeekWeek1, Settings: testCycleSettings()},
	}
	handler := NewHandler(store, nil)

	res, err := handler.ActivateCycle(authContext(store.user.ID), api.ActivateCycleParams{CycleId: secondID})
	if err != nil {
		t.Fatalf("ActivateCycle() error = %v", err)
	}
	activated, ok := res.(*api.CycleResponse)
	if !ok {
		t.Fatalf("ActivateCycle() response = %T, want *api.CycleResponse", res)
	}
	if activated.Cycle.ID != secondID || activated.Cycle.Status != api.CycleStatusActive {
		t.Fatalf("activated cycle = %#v", activated.Cycle)
	}
	if store.cycles[0].Status != api.CycleStatusArchived {
		t.Fatalf("previous status = %s, want archived", store.cycles[0].Status)
	}
}

func TestGetExerciseDetailsFallsBackToCatalogMedia(t *testing.T) {
	store := newFakeStore()
	store.exerciseDetails = map[string]api.ExerciseDetails{
		"reverse_grip_bench": {
			ExerciseKey: "reverse_grip_bench",
			Name:        "Жим обратным хватом",
			AliasStatus: api.ExerciseDetailsAliasStatusNeedsReview,
			Media:       api.ExerciseMedia{Status: api.ExerciseMediaStatusMissing},
		},
	}
	manifestPath := filepath.Join(t.TempDir(), "exercise_media.json")
	if err := os.WriteFile(manifestPath, []byte(`[{"datasetExerciseId":"2187","gifUrl":"https://example.com/exercises/2187.gif"}]`), 0o600); err != nil {
		t.Fatalf("write media manifest: %v", err)
	}
	catalog, err := exercises.NewCatalog("", manifestPath)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	handler := NewHandler(store, catalog)

	res, err := handler.GetExerciseDetails(context.Background(), api.GetExerciseDetailsParams{ExerciseKey: "reverse_grip_bench"})
	if err != nil {
		t.Fatalf("GetExerciseDetails() error = %v", err)
	}
	body, ok := res.(*api.ExerciseDetailsResponse)
	if !ok {
		t.Fatalf("GetExerciseDetails() response = %T, want *api.ExerciseDetailsResponse", res)
	}
	if got := body.Exercise.DatasetExerciseId.Or(""); got != "2187" {
		t.Fatalf("dataset id = %q, want 2187", got)
	}
	if body.Exercise.Media.Status != api.ExerciseMediaStatusAvailable {
		t.Fatalf("media status = %q, want available", body.Exercise.Media.Status)
	}
}

func TestListExercisesAddsCodeAliasMatches(t *testing.T) {
	store := newFakeStore()
	store.user = api.User{ID: uuid.New(), Nickname: "athlete_1", CreatedAt: time.Now()}
	store.catalog = []api.ExerciseCatalogItem{
		{
			DatasetExerciseId: "2187",
			Name:              "barbell reverse close-grip bench press",
			Media:             api.ExerciseMedia{Status: api.ExerciseMediaStatusAvailable},
		},
	}
	catalog, err := exercises.NewCatalog("", "")
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	handler := NewHandler(store, catalog)

	res, err := handler.ListExercises(authContext(store.user.ID), api.ListExercisesParams{
		Query:  api.NewOptString("жим обратным хватом"),
		Limit:  api.NewOptInt(10),
		HasGif: api.NewOptBool(true),
	})
	if err != nil {
		t.Fatalf("ListExercises() error = %v", err)
	}
	body, ok := res.(*api.ExerciseCatalogListResponse)
	if !ok {
		t.Fatalf("ListExercises() response = %T, want *api.ExerciseCatalogListResponse", res)
	}
	if len(body.Items) != 1 || body.Items[0].DatasetExerciseId != "2187" {
		t.Fatalf("items = %#v, want only dataset 2187", body.Items)
	}
}

func authContext(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), userContextKey{}, userID)
}

func testCycleSettings() api.CycleSettings {
	return api.CycleSettings{
		OneRepMaxKg:     api.OneRepMaxes{Deadlift: 225, Bench: 125, Squat: 170},
		Variant:         api.ProgramVariantVariant1,
		ProgressionStep: api.ProgressionStepStep4Percent,
		Assistance: api.AssistanceSelection{
			Deadlift: "good_morning",
			Bench:    "close_grip_bench",
			Squat:    "front_squat",
		},
		Gpp: api.GPPSelection{
			Abs:            api.NewNilString("abs"),
			Triceps:        api.NewNilString("triceps"),
			HorizontalPull: api.NewNilString("barbell_row"),
			Biceps:         api.NewNilString("biceps"),
			VerticalPull:   api.NewNilString("pull_up"),
			OverheadPress:  api.NewNilString("barbell_military_press"),
		},
	}
}

type fakeStore struct {
	user             api.User
	passwordHash     string
	sessionTokenHash string
	sessions         map[string]uuid.UUID
	cycles           []api.ProgramCycle
	catalog          []api.ExerciseCatalogItem
	exerciseDetails  map[string]api.ExerciseDetails
}

func newFakeStore() *fakeStore {
	return &fakeStore{}
}

func (s *fakeStore) CreateUser(ctx context.Context, nickname, passwordHash string) (api.User, error) {
	s.user = api.User{ID: uuid.New(), Nickname: api.Nickname(nickname), CreatedAt: time.Now()}
	s.passwordHash = passwordHash
	return s.user, nil
}

func (s *fakeStore) UserByNickname(ctx context.Context, nickname string) (api.User, string, bool, error) {
	if string(s.user.Nickname) != nickname {
		return api.User{}, "", false, nil
	}
	return s.user, s.passwordHash, true, nil
}

func (s *fakeStore) UserByID(ctx context.Context, id uuid.UUID) (api.User, bool, error) {
	if s.user.ID != id {
		return api.User{}, false, nil
	}
	return s.user, true, nil
}

func (s *fakeStore) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	s.sessionTokenHash = tokenHash
	if s.sessions == nil {
		s.sessions = map[string]uuid.UUID{}
	}
	s.sessions[tokenHash] = userID
	return nil
}

func (s *fakeStore) Session(ctx context.Context, tokenHash string) (Session, bool, error) {
	if s.sessions != nil {
		userID, ok := s.sessions[tokenHash]
		if !ok {
			return Session{}, false, nil
		}
		return Session{UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, true, nil
	}
	return Session{UserID: s.user.ID, ExpiresAt: time.Now().Add(time.Hour)}, s.user.ID != uuid.Nil, nil
}

func (s *fakeStore) RevokeSession(ctx context.Context, tokenHash string) error {
	return nil
}

func (s *fakeStore) Profile(ctx context.Context, userID uuid.UUID) (api.AthleteProfile, error) {
	return api.AthleteProfile{}, nil
}

func (s *fakeStore) SaveProfile(ctx context.Context, userID uuid.UUID, input api.AthleteProfileInput) (api.AthleteProfile, error) {
	return api.AthleteProfile{}, nil
}

func (s *fakeStore) ListCycles(ctx context.Context, userID uuid.UUID) ([]api.ProgramCycle, uuid.UUID, bool, error) {
	for _, cycle := range s.cycles {
		if cycle.Status == api.CycleStatusActive {
			return append([]api.ProgramCycle(nil), s.cycles...), cycle.ID, true, nil
		}
	}
	return append([]api.ProgramCycle(nil), s.cycles...), uuid.UUID{}, false, nil
}

func (s *fakeStore) CreateCycle(ctx context.Context, userID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, error) {
	for i := range s.cycles {
		if s.cycles[i].Status == api.CycleStatusActive {
			s.cycles[i].Status = api.CycleStatusArchived
		}
	}
	cycle := api.ProgramCycle{
		ID:          uuid.New(),
		Title:       title,
		Status:      api.CycleStatusActive,
		CurrentWeek: week,
		Settings:    settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.cycles = append([]api.ProgramCycle{cycle}, s.cycles...)
	return cycle, nil
}

func (s *fakeStore) CurrentCycle(ctx context.Context, userID uuid.UUID) (api.ProgramCycle, bool, error) {
	for _, cycle := range s.cycles {
		if cycle.Status == api.CycleStatusActive {
			return cycle, true, nil
		}
	}
	return api.ProgramCycle{}, false, nil
}

func (s *fakeStore) SaveCurrentCycle(ctx context.Context, userID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, error) {
	return api.ProgramCycle{}, nil
}

func (s *fakeStore) SaveCycle(ctx context.Context, userID, cycleID uuid.UUID, title string, week api.ProgramWeek, settings api.CycleSettings) (api.ProgramCycle, bool, error) {
	for i := range s.cycles {
		if s.cycles[i].ID == cycleID {
			s.cycles[i].Title = title
			s.cycles[i].CurrentWeek = week
			s.cycles[i].Settings = settings
			return s.cycles[i], true, nil
		}
	}
	return api.ProgramCycle{}, false, nil
}

func (s *fakeStore) ActivateCycle(ctx context.Context, userID, cycleID uuid.UUID) (api.ProgramCycle, bool, error) {
	active := -1
	for i := range s.cycles {
		if s.cycles[i].ID == cycleID {
			active = i
			continue
		}
		if s.cycles[i].Status == api.CycleStatusActive {
			s.cycles[i].Status = api.CycleStatusArchived
		}
	}
	if active == -1 {
		return api.ProgramCycle{}, false, nil
	}
	s.cycles[active].Status = api.CycleStatusActive
	return s.cycles[active], true, nil
}

func (s *fakeStore) AdvanceCycle(ctx context.Context, userID uuid.UUID, week api.ProgramWeek) (api.ProgramCycle, bool, error) {
	return api.ProgramCycle{}, false, nil
}

func (s *fakeStore) ListProgress(ctx context.Context, cycleID uuid.UUID, week api.ProgramWeek) ([]api.ProgressCheckpoint, error) {
	return nil, nil
}

func (s *fakeStore) UpsertProgress(ctx context.Context, cycleID uuid.UUID, input api.ProgressCheckpointInput) (api.ProgressCheckpoint, error) {
	return api.ProgressCheckpoint{}, nil
}

func (s *fakeStore) DeleteProgress(ctx context.Context, cycleID, checkpointID uuid.UUID) (bool, error) {
	return false, nil
}

func (s *fakeStore) ExerciseDetails(ctx context.Context, exerciseKey string) (api.ExerciseDetails, bool, error) {
	if s.exerciseDetails != nil {
		details, ok := s.exerciseDetails[exerciseKey]
		return details, ok, nil
	}
	return api.ExerciseDetails{}, false, nil
}

func (s *fakeStore) ListExercises(ctx context.Context, params api.ListExercisesParams) (api.ExerciseCatalogListResponse, error) {
	return api.ExerciseCatalogListResponse{Items: append([]api.ExerciseCatalogItem(nil), s.catalog...), Total: len(s.catalog), Limit: 30}, nil
}

func (s *fakeStore) CatalogExercise(ctx context.Context, datasetExerciseID string) (api.ExerciseCatalogItem, bool, error) {
	for _, item := range s.catalog {
		if item.DatasetExerciseId == datasetExerciseID {
			return item, true, nil
		}
	}
	return api.ExerciseCatalogItem{}, false, nil
}
