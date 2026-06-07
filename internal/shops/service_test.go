package shops

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shophub-platform/shophub/internal/models"
)

// --- testify mock-ovi za Repository i Orchestrator ---

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, shop *models.Shop) error {
	return m.Called(ctx, shop).Error(0)
}
func (m *mockRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Shop, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]models.Shop), args.Error(1)
}
func (m *mockRepo) GetByID(ctx context.Context, ownerID, id uuid.UUID) (*models.Shop, error) {
	args := m.Called(ctx, ownerID, id)
	shop, _ := args.Get(0).(*models.Shop)
	return shop, args.Error(1)
}
func (m *mockRepo) Update(ctx context.Context, shop *models.Shop) error {
	return m.Called(ctx, shop).Error(0)
}
func (m *mockRepo) Delete(ctx context.Context, ownerID, id uuid.UUID) error {
	return m.Called(ctx, ownerID, id).Error(0)
}

type mockOrch struct{ mock.Mock }

func (m *mockOrch) Apply(ctx context.Context, r ShopResource) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mockOrch) Status(ctx context.Context, ns, name string) (ResourceStatus, error) {
	args := m.Called(ctx, ns, name)
	return args.Get(0).(ResourceStatus), args.Error(1)
}
func (m *mockOrch) Delete(ctx context.Context, ns, name string) error {
	return m.Called(ctx, ns, name).Error(0)
}

func newSvc(repo Repository, orch Orchestrator) Service {
	return NewService(repo, orch, Options{Namespace: "default", DefaultImage: "img:1", URLTemplate: "http://%s.test"})
}

func TestCreateSuccess(t *testing.T) {
	repo := &mockRepo{}
	orch := &mockOrch{}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Shop")).Return(nil)
	orch.On("Apply", mock.Anything, mock.AnythingOfType("shops.ShopResource")).Return(nil)
	orch.On("Status", mock.Anything, "default", mock.Anything).Return(ResourceStatus{Phase: "Ready"}, nil)

	svc := newSvc(repo, orch)
	v, err := svc.Create(context.Background(), uuid.New(), CreateInput{
		Name: "moja-radnja", Availability: "high", WalletAddress: "0xabc", DatabaseType: "postgres",
	})
	require.NoError(t, err)
	assert.Equal(t, "running", v.Status)
	assert.Equal(t, "moja-radnja", v.Shop.Name)
	assert.Equal(t, 3, v.Shop.Replicas())
	repo.AssertExpectations(t)
	orch.AssertExpectations(t)
}

func TestCreateValidation(t *testing.T) {
	svc := newSvc(&mockRepo{}, &mockOrch{})
	_, err := svc.Create(context.Background(), uuid.New(), CreateInput{Name: "  ", WalletAddress: "0x"})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateBadAvailability(t *testing.T) {
	svc := newSvc(&mockRepo{}, &mockOrch{})
	_, err := svc.Create(context.Background(), uuid.New(), CreateInput{Name: "x", Availability: "mega", WalletAddress: "0x"})
	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateOrchestratorFailsNoDBWrite(t *testing.T) {
	repo := &mockRepo{}
	orch := &mockOrch{}
	orch.On("Apply", mock.Anything, mock.Anything).Return(errors.New("api server down"))

	svc := newSvc(repo, orch)
	_, err := svc.Create(context.Background(), uuid.New(), CreateInput{Name: "x", WalletAddress: "0xabc"})
	require.Error(t, err)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateDBFailsCleansUpCR(t *testing.T) {
	repo := &mockRepo{}
	orch := &mockOrch{}
	orch.On("Apply", mock.Anything, mock.Anything).Return(nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("unique violation"))
	orch.On("Delete", mock.Anything, "default", mock.Anything).Return(nil)

	svc := newSvc(repo, orch)
	_, err := svc.Create(context.Background(), uuid.New(), CreateInput{Name: "x", WalletAddress: "0xabc"})
	require.Error(t, err)
	orch.AssertCalled(t, "Delete", mock.Anything, "default", mock.Anything)
}

func TestGetNotFound(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return((*models.Shop)(nil), ErrNotFound)
	svc := newSvc(repo, &mockOrch{})
	_, err := svc.Get(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestListMapsStatus(t *testing.T) {
	owner := uuid.New()
	shop := models.Shop{ID: uuid.New(), OwnerID: owner, Name: "a", Availability: "standard", DatabaseType: "postgres"}
	repo := &mockRepo{}
	orch := &mockOrch{}
	repo.On("ListByOwner", mock.Anything, owner).Return([]models.Shop{shop}, nil)
	orch.On("Status", mock.Anything, "default", CRName(shop.ID)).Return(ResourceStatus{Phase: "Provisioning"}, nil)

	svc := newSvc(repo, orch)
	views, err := svc.List(context.Background(), owner)
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Equal(t, "deploying", views[0].Status)
}

func TestUpdateChangesAvailability(t *testing.T) {
	owner := uuid.New()
	shop := &models.Shop{ID: uuid.New(), OwnerID: owner, Name: "a", Availability: "standard", DatabaseType: "postgres", WalletAddr: "0x"}
	repo := &mockRepo{}
	orch := &mockOrch{}
	repo.On("GetByID", mock.Anything, owner, shop.ID).Return(shop, nil)
	orch.On("Apply", mock.Anything, mock.Anything).Return(nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	orch.On("Status", mock.Anything, "default", mock.Anything).Return(ResourceStatus{Phase: "Ready"}, nil)

	svc := newSvc(repo, orch)
	high := "high"
	v, err := svc.Update(context.Background(), owner, shop.ID, UpdateInput{Availability: &high})
	require.NoError(t, err)
	assert.Equal(t, models.AvailabilityHigh, v.Shop.Availability)
}

func TestDeleteOwnershipChecked(t *testing.T) {
	owner := uuid.New()
	id := uuid.New()
	repo := &mockRepo{}
	orch := &mockOrch{}
	repo.On("GetByID", mock.Anything, owner, id).Return(&models.Shop{ID: id, OwnerID: owner}, nil)
	orch.On("Delete", mock.Anything, "default", CRName(id)).Return(nil)
	repo.On("Delete", mock.Anything, owner, id).Return(nil)

	svc := newSvc(repo, orch)
	require.NoError(t, svc.Delete(context.Background(), owner, id))
	orch.AssertExpectations(t)
}

func TestURL(t *testing.T) {
	owner := uuid.New()
	id := uuid.New()
	repo := &mockRepo{}
	repo.On("GetByID", mock.Anything, owner, id).Return(&models.Shop{ID: id, OwnerID: owner}, nil)
	svc := newSvc(repo, &mockOrch{})
	url, err := svc.URL(context.Background(), owner, id)
	require.NoError(t, err)
	assert.Equal(t, "http://"+CRName(id)+".test", url)
}

func TestPhaseToStatus(t *testing.T) {
	assert.Equal(t, "running", phaseToStatus("Ready"))
	assert.Equal(t, "deploying", phaseToStatus("Provisioning"))
	assert.Equal(t, "error", phaseToStatus("Failed"))
	assert.Equal(t, "pending", phaseToStatus(""))
	assert.Equal(t, "pending", phaseToStatus("Pending"))
}
