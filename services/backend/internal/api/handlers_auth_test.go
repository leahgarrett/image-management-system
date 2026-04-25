package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leahgarrett/image-management-system/services/backend/internal/api"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
	"github.com/leahgarrett/image-management-system/services/backend/internal/mailer"
)

// mockQuerier implements db.Querier for tests. All methods not under test
// return zero values or nil errors by default.
type mockQuerier struct {
	// Auth fields
	getUserByEmailFn          func(ctx context.Context, email string) (db.User, error)
	getUserByIDFn             func(ctx context.Context, id uuid.UUID) (db.User, error)
	listUsersFn               func(ctx context.Context) ([]db.User, error)
	createUserFn              func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	updateUserLastLoginFn     func(ctx context.Context, id uuid.UUID) error
	createMagicLinkTokenFn    func(ctx context.Context, arg db.CreateMagicLinkTokenParams) (db.MagicLinkToken, error)
	getMagicLinkTokenByHashFn func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error)
	markTokenUsedFn           func(ctx context.Context, id uuid.UUID) error

	// Image fields
	createImageFn       func(ctx context.Context, arg db.CreateImageParams) (db.Image, error)
	getImageByIDFn      func(ctx context.Context, id uuid.UUID) (db.Image, error)
	getImageByImageIDFn func(ctx context.Context, imageID string) (db.Image, error)
	listImagesFn        func(ctx context.Context, arg db.ListImagesParams) ([]db.Image, error)
	updateImageFn       func(ctx context.Context, arg db.UpdateImageParams) (db.Image, error)
	deleteImageFn       func(ctx context.Context, id uuid.UUID) error
	createImagePersonFn func(ctx context.Context, arg db.CreateImagePersonParams) (db.ImagePerson, error)
	deleteImagePeopleFn func(ctx context.Context, imageID uuid.UUID) error
	listImagePeopleFn   func(ctx context.Context, imageID uuid.UUID) ([]db.ImagePerson, error)

	// Tag fields
	createTagFn         func(ctx context.Context, arg db.CreateTagParams) (db.Tag, error)
	listTagsFn          func(ctx context.Context) ([]db.Tag, error)
	searchTagsFn        func(ctx context.Context, query string) ([]db.Tag, error)
	addImageTagFn       func(ctx context.Context, arg db.AddImageTagParams) error
	removeImageTagFn    func(ctx context.Context, arg db.RemoveImageTagParams) error
	listImageTagsFn     func(ctx context.Context, imageID uuid.UUID) ([]db.Tag, error)
	incrementTagUsageFn func(ctx context.Context, id uuid.UUID) error
	decrementTagUsageFn func(ctx context.Context, id uuid.UUID) error

	// User admin fields
	updateUserRoleFn   func(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error)
	updateUserStatusFn func(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error)
}

func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return db.User{}, nil
}
func (m *mockQuerier) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return db.User{}, nil
}
func (m *mockQuerier) ListUsers(ctx context.Context) ([]db.User, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx)
	}
	return []db.User{}, nil
}
func (m *mockQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, arg)
	}
	return db.User{Email: arg.Email, Role: arg.Role, Status: arg.Status}, nil
}
func (m *mockQuerier) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	if m.updateUserLastLoginFn != nil {
		return m.updateUserLastLoginFn(ctx, id)
	}
	return nil
}
func (m *mockQuerier) CreateMagicLinkToken(ctx context.Context, arg db.CreateMagicLinkTokenParams) (db.MagicLinkToken, error) {
	if m.createMagicLinkTokenFn != nil {
		return m.createMagicLinkTokenFn(ctx, arg)
	}
	return db.MagicLinkToken{TokenHash: arg.TokenHash}, nil
}
func (m *mockQuerier) GetMagicLinkTokenByHash(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
	if m.getMagicLinkTokenByHashFn != nil {
		return m.getMagicLinkTokenByHashFn(ctx, tokenHash)
	}
	return db.MagicLinkToken{}, nil
}
func (m *mockQuerier) MarkTokenUsed(ctx context.Context, id uuid.UUID) error {
	if m.markTokenUsedFn != nil {
		return m.markTokenUsedFn(ctx, id)
	}
	return nil
}
func (m *mockQuerier) CreateImage(ctx context.Context, arg db.CreateImageParams) (db.Image, error) {
	if m.createImageFn != nil {
		return m.createImageFn(ctx, arg)
	}
	return db.Image{ImageID: arg.ImageID}, nil
}
func (m *mockQuerier) GetImageByID(ctx context.Context, id uuid.UUID) (db.Image, error) {
	if m.getImageByIDFn != nil {
		return m.getImageByIDFn(ctx, id)
	}
	return db.Image{}, nil
}
func (m *mockQuerier) GetImageByImageID(ctx context.Context, imageID string) (db.Image, error) {
	if m.getImageByImageIDFn != nil {
		return m.getImageByImageIDFn(ctx, imageID)
	}
	return db.Image{}, nil
}
func (m *mockQuerier) ListImages(ctx context.Context, arg db.ListImagesParams) ([]db.Image, error) {
	if m.listImagesFn != nil {
		return m.listImagesFn(ctx, arg)
	}
	return []db.Image{}, nil
}
func (m *mockQuerier) UpdateImage(ctx context.Context, arg db.UpdateImageParams) (db.Image, error) {
	if m.updateImageFn != nil {
		return m.updateImageFn(ctx, arg)
	}
	return db.Image{}, nil
}
func (m *mockQuerier) DeleteImage(ctx context.Context, id uuid.UUID) error {
	if m.deleteImageFn != nil {
		return m.deleteImageFn(ctx, id)
	}
	return nil
}
func (m *mockQuerier) CreateImagePerson(ctx context.Context, arg db.CreateImagePersonParams) (db.ImagePerson, error) {
	if m.createImagePersonFn != nil {
		return m.createImagePersonFn(ctx, arg)
	}
	return db.ImagePerson{Name: arg.Name}, nil
}
func (m *mockQuerier) DeleteImagePeople(ctx context.Context, imageID uuid.UUID) error {
	if m.deleteImagePeopleFn != nil {
		return m.deleteImagePeopleFn(ctx, imageID)
	}
	return nil
}
func (m *mockQuerier) ListImagePeople(ctx context.Context, imageID uuid.UUID) ([]db.ImagePerson, error) {
	if m.listImagePeopleFn != nil {
		return m.listImagePeopleFn(ctx, imageID)
	}
	return []db.ImagePerson{}, nil
}
func (m *mockQuerier) CreateTag(ctx context.Context, arg db.CreateTagParams) (db.Tag, error) {
	if m.createTagFn != nil {
		return m.createTagFn(ctx, arg)
	}
	return db.Tag{Name: arg.Name}, nil
}
func (m *mockQuerier) ListTags(ctx context.Context) ([]db.Tag, error) {
	if m.listTagsFn != nil {
		return m.listTagsFn(ctx)
	}
	return []db.Tag{}, nil
}
func (m *mockQuerier) SearchTags(ctx context.Context, query string) ([]db.Tag, error) {
	if m.searchTagsFn != nil {
		return m.searchTagsFn(ctx, query)
	}
	return []db.Tag{}, nil
}
func (m *mockQuerier) AddImageTag(ctx context.Context, arg db.AddImageTagParams) error {
	if m.addImageTagFn != nil {
		return m.addImageTagFn(ctx, arg)
	}
	return nil
}
func (m *mockQuerier) RemoveImageTag(ctx context.Context, arg db.RemoveImageTagParams) error {
	if m.removeImageTagFn != nil {
		return m.removeImageTagFn(ctx, arg)
	}
	return nil
}
func (m *mockQuerier) ListImageTags(ctx context.Context, imageID uuid.UUID) ([]db.Tag, error) {
	if m.listImageTagsFn != nil {
		return m.listImageTagsFn(ctx, imageID)
	}
	return []db.Tag{}, nil
}
func (m *mockQuerier) IncrementTagUsage(ctx context.Context, id uuid.UUID) error {
	if m.incrementTagUsageFn != nil {
		return m.incrementTagUsageFn(ctx, id)
	}
	return nil
}
func (m *mockQuerier) DecrementTagUsage(ctx context.Context, id uuid.UUID) error {
	if m.decrementTagUsageFn != nil {
		return m.decrementTagUsageFn(ctx, id)
	}
	return nil
}
func (m *mockQuerier) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	if m.updateUserRoleFn != nil {
		return m.updateUserRoleFn(ctx, arg)
	}
	return db.User{}, nil
}
func (m *mockQuerier) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	if m.updateUserStatusFn != nil {
		return m.updateUserStatusFn(ctx, arg)
	}
	return db.User{}, nil
}

// --- Auth handler tests ---

func newTestHandlers(q db.Querier) *api.Handlers {
	return api.NewHandlers(q, mailer.NewLogMailer(), "http://localhost:3000", testSecret, true)
}

func TestLogin_CreatesTokenAndSendsLink(t *testing.T) {
	q := &mockQuerier{
		getUserByEmailFn: func(ctx context.Context, email string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
	}
	h := newTestHandlers(q)
	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["message"] != "Magic link sent" {
		t.Errorf("unexpected message: %s", resp["message"])
	}
}

func TestVerify_ValidToken_SetsCookie(t *testing.T) {
	validUUID := uuid.New()
	q := &mockQuerier{
		getMagicLinkTokenByHashFn: func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
			return db.MagicLinkToken{
				ID:        validUUID,
				UserID:    validUUID,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(time.Minute),
			}, nil
		},
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (db.User, error) {
			return db.User{ID: validUUID, Email: "test@example.com", Role: "contributor"}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify?token=abc123deadbeef", nil)
	rr := httptest.NewRecorder()
	h.Verify(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "auth_token" && c.HttpOnly {
			found = true
		}
	}
	if !found {
		t.Error("expected httpOnly auth_token cookie to be set")
	}
}

func TestVerify_ExpiredToken_Returns401(t *testing.T) {
	validUUID := uuid.New()
	q := &mockQuerier{
		getMagicLinkTokenByHashFn: func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
			return db.MagicLinkToken{
				ID:        validUUID,
				UserID:    validUUID,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(-time.Minute),
			}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify?token=abc123", nil)
	rr := httptest.NewRecorder()
	h.Verify(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestVerify_UsedToken_Returns401(t *testing.T) {
	validUUID := uuid.New()
	q := &mockQuerier{
		getMagicLinkTokenByHashFn: func(ctx context.Context, tokenHash string) (db.MagicLinkToken, error) {
			return db.MagicLinkToken{
				ID:        validUUID,
				UserID:    validUUID,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(time.Minute),
				UsedAt:    sql.NullTime{Time: time.Now().Add(-time.Minute), Valid: true},
			}, nil
		},
	}
	h := newTestHandlers(q)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify?token=abc123", nil)
	rr := httptest.NewRecorder()
	h.Verify(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestLogout_ClearsCookie(t *testing.T) {
	h := newTestHandlers(&mockQuerier{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rr := httptest.NewRecorder()
	h.Logout(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "auth_token" && c.MaxAge != 0 {
			t.Errorf("expected MaxAge=0 for cleared cookie, got %d", c.MaxAge)
		}
	}
}
