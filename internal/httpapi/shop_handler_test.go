package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shophub-platform/shophub/internal/auth"
	"github.com/shophub-platform/shophub/internal/models"
	"github.com/shophub-platform/shophub/internal/shops"
)

// stubTokens je lažni authTokenManager: token "bad" je nevalidan, sve ostalo
// vraća claims za fiksnog korisnika.
type stubTokens struct{ uid uuid.UUID }

func (s stubTokens) Parse(tok string, _ auth.TokenType) (*auth.Claims, error) {
	if tok == "bad" {
		return nil, auth.ErrInvalidToken
	}
	return &auth.Claims{UserID: s.uid, Email: "a@b.com", Type: auth.AccessToken}, nil
}

// mockService je testify mock shops.Service interfejsa.
type mockService struct{ mock.Mock }

func (m *mockService) List(ctx context.Context, owner uuid.UUID) ([]shops.View, error) {
	args := m.Called(ctx, owner)
	return args.Get(0).([]shops.View), args.Error(1)
}
func (m *mockService) Create(ctx context.Context, owner uuid.UUID, in shops.CreateInput) (*shops.View, error) {
	args := m.Called(ctx, owner, in)
	v, _ := args.Get(0).(*shops.View)
	return v, args.Error(1)
}
func (m *mockService) Get(ctx context.Context, owner, id uuid.UUID) (*shops.View, error) {
	args := m.Called(ctx, owner, id)
	v, _ := args.Get(0).(*shops.View)
	return v, args.Error(1)
}
func (m *mockService) Update(ctx context.Context, owner, id uuid.UUID, in shops.UpdateInput) (*shops.View, error) {
	args := m.Called(ctx, owner, id, in)
	v, _ := args.Get(0).(*shops.View)
	return v, args.Error(1)
}
func (m *mockService) Delete(ctx context.Context, owner, id uuid.UUID) error {
	return m.Called(ctx, owner, id).Error(0)
}
func (m *mockService) URL(ctx context.Context, owner, id uuid.UUID) (string, error) {
	args := m.Called(ctx, owner, id)
	return args.String(0), args.Error(1)
}

func testServer(svc shops.Service, uid uuid.UUID) http.Handler {
	h := NewShopHandler(svc, stubTokens{uid: uid})
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func do(t *testing.T, srv http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func TestCreateHandlerOK(t *testing.T) {
	uid := uuid.New()
	id := uuid.New()
	svc := &mockService{}
	svc.On("Create", mock.Anything, uid, mock.AnythingOfType("shops.CreateInput")).
		Return(&shops.View{Shop: models.Shop{ID: id, Name: "x"}, Status: "pending", URL: "http://x.test"}, nil)

	srv := testServer(svc, uid)
	w := do(t, srv, "POST", "/api/v1/shops", `{"name":"x","availability":"standard","walletAddress":"0x","databaseType":"postgres"}`, "good")
	require.Equal(t, http.StatusCreated, w.Code)

	var resp shopResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, id.String(), resp.ID)
	assert.Equal(t, "pending", resp.Status)
}

func TestCreateHandlerValidation(t *testing.T) {
	uid := uuid.New()
	svc := &mockService{}
	svc.On("Create", mock.Anything, uid, mock.Anything).Return((*shops.View)(nil), shops.ErrValidation)
	srv := testServer(svc, uid)
	w := do(t, srv, "POST", "/api/v1/shops", `{"name":""}`, "good")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthRequired(t *testing.T) {
	srv := testServer(&mockService{}, uuid.New())
	w := do(t, srv, "GET", "/api/v1/shops", "", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = do(t, srv, "GET", "/api/v1/shops", "", "bad")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetNotFoundHandler(t *testing.T) {
	uid := uuid.New()
	id := uuid.New()
	svc := &mockService{}
	svc.On("Get", mock.Anything, uid, id).Return((*shops.View)(nil), shops.ErrNotFound)
	srv := testServer(svc, uid)
	w := do(t, srv, "GET", "/api/v1/shops/"+id.String(), "", "good")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetBadID(t *testing.T) {
	uid := uuid.New()
	srv := testServer(&mockService{}, uid)
	w := do(t, srv, "GET", "/api/v1/shops/not-a-uuid", "", "good")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteHandler(t *testing.T) {
	uid := uuid.New()
	id := uuid.New()
	svc := &mockService{}
	svc.On("Delete", mock.Anything, uid, id).Return(nil)
	srv := testServer(svc, uid)
	w := do(t, srv, "DELETE", "/api/v1/shops/"+id.String(), "", "good")
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestListHandler(t *testing.T) {
	uid := uuid.New()
	svc := &mockService{}
	svc.On("List", mock.Anything, uid).Return([]shops.View{
		{Shop: models.Shop{ID: uuid.New(), Name: "a"}, Status: "running", URL: "http://a.test"},
	}, nil)
	srv := testServer(svc, uid)
	w := do(t, srv, "GET", "/api/v1/shops", "", "good")
	require.Equal(t, http.StatusOK, w.Code)
	var resp []shopResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
	assert.Equal(t, "running", resp[0].Status)
}

func TestURLHandler(t *testing.T) {
	uid := uuid.New()
	id := uuid.New()
	svc := &mockService{}
	svc.On("URL", mock.Anything, uid, id).Return("http://shop-x.test", nil)
	srv := testServer(svc, uid)
	w := do(t, srv, "GET", "/api/v1/shops/"+id.String()+"/url", "", "good")
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "http://shop-x.test", resp["url"])
}

func TestUpdateHandler(t *testing.T) {
	uid := uuid.New()
	id := uuid.New()
	svc := &mockService{}
	svc.On("Update", mock.Anything, uid, id, mock.AnythingOfType("shops.UpdateInput")).
		Return(&shops.View{Shop: models.Shop{ID: id, Availability: "high"}, Status: "deploying"}, nil)
	srv := testServer(svc, uid)
	w := do(t, srv, "PATCH", "/api/v1/shops/"+id.String(), `{"availability":"high"}`, "good")
	require.Equal(t, http.StatusOK, w.Code)
	var resp shopResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "high", resp.Availability)
}
