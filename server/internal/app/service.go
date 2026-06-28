package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ogen-go/ogen/ogenerrors"
	"golang.org/x/crypto/bcrypt"

	"sport/server/internal/api"
	"sport/server/internal/exercises"
	"sport/server/internal/program"
)

const sessionTTL = 30 * 24 * time.Hour

var (
	errUnauthorized = errors.New("unauthorized")
	errNotFound     = errors.New("not found")
	errConflict     = errors.New("conflict")
)

type userContextKey struct{}

type Handler struct {
	store   Store
	catalog *exercises.Catalog
}

func NewHandler(store Store, catalog *exercises.Catalog) *Handler {
	return &Handler{store: store, catalog: catalog}
}

func (h *Handler) Healthz(ctx context.Context) (*api.HealthResponse, error) {
	return &api.HealthResponse{Status: api.HealthResponseStatusOk}, nil
}

func (h *Handler) Bootstrap(ctx context.Context) (*api.BootstrapResponseHeaders, error) {
	return &api.BootstrapResponseHeaders{
		SetCookie: api.NewOptString("init=1; Path=/; SameSite=Lax"),
		Response: api.BootstrapResponse{
			InitCookie: api.InitCookieInfo{
				Name:       "init",
				Value:      "1",
				JsReadable: true,
				Purpose:    "Temporary bootstrap marker. Not authentication.",
			},
			Service: api.ServiceInfo{
				AuthMode:    "local_nickname_password",
				HasDatabase: true,
			},
		},
	}, nil
}

func (h *Handler) Register(ctx context.Context, req *api.RegisterRequest) (api.RegisterRes, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := h.store.CreateUser(ctx, string(req.Nickname), string(hash))
	if errors.Is(err, errConflict) {
		return &api.RegisterConflict{Error: errorBody("conflict", "nickname already exists")}, nil
	}
	if err != nil {
		return nil, err
	}
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	if err := h.store.CreateSession(ctx, user.ID, hashToken(token), time.Now().Add(sessionTTL)); err != nil {
		return nil, err
	}
	return authHeaders(user, token), nil
}

func (h *Handler) Login(ctx context.Context, req *api.LoginRequest) (api.LoginRes, error) {
	user, passwordHash, ok, err := h.store.UserByNickname(ctx, string(req.Nickname))
	if err != nil {
		return nil, err
	}
	if !ok || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		return &api.LoginUnauthorized{Error: errorBody("unauthorized", "invalid nickname or password")}, nil
	}
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	if err := h.store.CreateSession(ctx, user.ID, hashToken(token), time.Now().Add(sessionTTL)); err != nil {
		return nil, err
	}
	return authHeaders(user, token), nil
}

func (h *Handler) Logout(ctx context.Context) (api.LogoutRes, error) {
	if sessionHash, ok := ctx.Value(sessionHashContextKey{}).(string); ok {
		if err := h.store.RevokeSession(ctx, sessionHash); err != nil {
			return nil, err
		}
	}
	return &api.LogoutNoContent{SetCookie: api.NewOptString("sid=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0")}, nil
}

func (h *Handler) GetMe(ctx context.Context) (api.GetMeRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return unauthorized(), nil
	}
	return &api.MeResponse{User: user}, nil
}

func (h *Handler) GetMyProfile(ctx context.Context) (api.GetMyProfileRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return unauthorized(), nil
	}
	profile, err := h.store.Profile(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &api.AthleteProfileResponse{Profile: profile}, nil
}

func (h *Handler) PutMyProfile(ctx context.Context, req *api.PutAthleteProfileRequest) (api.PutMyProfileRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.PutMyProfileUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	if err := validateProfile(req.Profile); err != nil {
		return &api.PutMyProfileBadRequest{Error: errorBody("validation_error", err.Error())}, nil
	}
	profile, err := h.store.SaveProfile(ctx, user.ID, req.Profile)
	if err != nil {
		return nil, err
	}
	return &api.AthleteProfileResponse{Profile: profile}, nil
}

func (h *Handler) GetProgramOptions(ctx context.Context) (*api.ProgramOptionsResponse, error) {
	return program.Options(), nil
}

func (h *Handler) CalculateProgram(ctx context.Context, req *api.CalculateProgramRequest) (api.CalculateProgramRes, error) {
	plan, err := program.Calculate(req.Selection)
	if err != nil {
		return &api.ErrorResponse{Error: errorBody("validation_error", err.Error())}, nil
	}
	return plan, nil
}

func (h *Handler) GetCurrentCycle(ctx context.Context) (api.GetCurrentCycleRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.GetCurrentCycleUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	cycle, ok, err := h.store.CurrentCycle(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &api.GetCurrentCycleNotFound{Error: errorBody("not_found", "active cycle not found")}, nil
	}
	return &api.CycleResponse{Cycle: cycle}, nil
}

func (h *Handler) PutCurrentCycle(ctx context.Context, req *api.PutCurrentCycleRequest) (api.PutCurrentCycleRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.PutCurrentCycleUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	if _, err := program.Calculate(api.ProgramSelection{Settings: req.Settings, Week: req.CurrentWeek}); err != nil {
		return &api.PutCurrentCycleBadRequest{Error: errorBody("validation_error", err.Error())}, nil
	}
	cycle, err := h.store.SaveCurrentCycle(ctx, user.ID, req.Title, req.CurrentWeek, req.Settings)
	if err != nil {
		return nil, err
	}
	return &api.CycleResponse{Cycle: cycle}, nil
}

func (h *Handler) AdvanceCurrentCycle(ctx context.Context, req *api.AdvanceCycleRequest) (api.AdvanceCurrentCycleRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.AdvanceCurrentCycleUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	cycle, ok, err := h.store.AdvanceCycle(ctx, user.ID, req.TargetWeek)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &api.AdvanceCurrentCycleNotFound{Error: errorBody("not_found", "active cycle not found")}, nil
	}
	return &api.CycleResponse{Cycle: cycle}, nil
}

func (h *Handler) ListCurrentCycleProgress(ctx context.Context, params api.ListCurrentCycleProgressParams) (api.ListCurrentCycleProgressRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.ListCurrentCycleProgressUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	cycle, ok, err := h.store.CurrentCycle(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &api.ListCurrentCycleProgressNotFound{Error: errorBody("not_found", "active cycle not found")}, nil
	}
	week := params.Week.Or("")
	items, err := h.store.ListProgress(ctx, cycle.ID, week)
	if err != nil {
		return nil, err
	}
	return &api.ProgressCheckpointsResponse{Items: items}, nil
}

func (h *Handler) UpsertCurrentCycleCheckpoint(ctx context.Context, req *api.UpsertProgressCheckpointRequest) (api.UpsertCurrentCycleCheckpointRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.UpsertCurrentCycleCheckpointUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	cycle, ok, err := h.store.CurrentCycle(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &api.UpsertCurrentCycleCheckpointNotFound{Error: errorBody("not_found", "active cycle not found")}, nil
	}
	if req.Checkpoint.ExerciseKey == "" {
		return &api.UpsertCurrentCycleCheckpointBadRequest{Error: errorBody("validation_error", "exerciseKey is required")}, nil
	}
	checkpoint, err := h.store.UpsertProgress(ctx, cycle.ID, req.Checkpoint)
	if err != nil {
		return nil, err
	}
	return &api.ProgressCheckpointResponse{Checkpoint: checkpoint}, nil
}

func (h *Handler) DeleteCurrentCycleCheckpoint(ctx context.Context, params api.DeleteCurrentCycleCheckpointParams) (api.DeleteCurrentCycleCheckpointRes, error) {
	user, err := h.currentUser(ctx)
	if err != nil {
		return &api.DeleteCurrentCycleCheckpointUnauthorized{Error: errorBody("unauthorized", "missing or invalid session")}, nil
	}
	cycle, ok, err := h.store.CurrentCycle(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	deleted := false
	if ok {
		deleted, err = h.store.DeleteProgress(ctx, cycle.ID, params.CheckpointId)
		if err != nil {
			return nil, err
		}
	}
	if !ok || !deleted {
		return &api.DeleteCurrentCycleCheckpointNotFound{Error: errorBody("not_found", "checkpoint not found")}, nil
	}
	return &api.DeleteCurrentCycleCheckpointNoContent{}, nil
}

func (h *Handler) GetExerciseDetails(ctx context.Context, params api.GetExerciseDetailsParams) (api.GetExerciseDetailsRes, error) {
	details, ok := h.catalog.Details(params.ExerciseKey)
	if !ok {
		return &api.ErrorResponse{Error: errorBody("not_found", "exercise not found")}, nil
	}
	return &api.ExerciseDetailsResponse{Exercise: details}, nil
}

func (h *Handler) currentUser(ctx context.Context) (api.User, error) {
	userID, ok := ctx.Value(userContextKey{}).(uuid.UUID)
	if !ok {
		return api.User{}, errUnauthorized
	}
	user, ok, err := h.store.UserByID(ctx, userID)
	if err != nil {
		return api.User{}, err
	}
	if !ok {
		return api.User{}, errUnauthorized
	}
	return user, nil
}

type sessionHashContextKey struct{}

type Security struct {
	store Store
}

func NewSecurity(store Store) *Security {
	return &Security{store: store}
}

func (s *Security) HandleSessionCookie(ctx context.Context, operationName api.OperationName, t api.SessionCookie) (context.Context, error) {
	sessionHash := hashToken(t.APIKey)
	session, ok, err := s.store.Session(ctx, sessionHash)
	if err != nil {
		return ctx, err
	}
	if !ok {
		return ctx, errUnauthorized
	}
	ctx = context.WithValue(ctx, userContextKey{}, session.UserID)
	ctx = context.WithValue(ctx, sessionHashContextKey{}, sessionHash)
	return ctx, nil
}

func ErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "internal server error"

	var securityErr *ogenerrors.SecurityError
	var decodeErr *ogenerrors.DecodeRequestError
	var decodeParamErr *ogenerrors.DecodeParamError
	switch {
	case errors.As(err, &securityErr):
		status = http.StatusUnauthorized
		code = "unauthorized"
		message = "missing or invalid session"
	case errors.As(err, &decodeErr), errors.As(err, &decodeParamErr):
		status = http.StatusBadRequest
		code = "validation_error"
		message = "request validation failed"
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(api.ErrorResponse{Error: errorBody(code, message)})
}

func authHeaders(user api.User, token string) *api.AuthResponseHeaders {
	return &api.AuthResponseHeaders{
		SetCookie: api.NewOptString("sid=" + token + "; Path=/; HttpOnly; SameSite=Lax"),
		Response:  api.AuthResponse{User: user},
	}
}

func validateProfile(profile api.AthleteProfileInput) error {
	for _, value := range []api.NilFloat64{profile.OneRepMaxKg.Deadlift, profile.OneRepMaxKg.Bench, profile.OneRepMaxKg.Squat} {
		if v, ok := value.Get(); ok && v <= 0 {
			return errors.New("profile one-rep max values must be positive when set")
		}
	}
	return nil
}

func unauthorized() *api.ErrorResponse {
	return &api.ErrorResponse{Error: errorBody("unauthorized", "missing or invalid session")}
}

func errorBody(code, message string) api.APIError {
	return api.APIError{Code: code, Message: message, Details: []string{}}
}

func newToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
