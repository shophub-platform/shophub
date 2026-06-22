// Package shops sadrži poslovnu logiku za upravljanje sajtovima prodavnica
// (FZ 2.1): CRUD nad Shop meta-podacima u bazi + sinhronizacija Shop CR-a
// u Kubernetes klasteru preko Orchestrator-a (client-go).
package shops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/shophub-platform/shophub/internal/metrics"
	"github.com/shophub-platform/shophub/internal/models"
)

// Greške koje servisni sloj vraća. Handler ih mapira na HTTP statuse.
var (
	// ErrNotFound se vraća kada prodavnica ne postoji ili ne pripada korisniku.
	ErrNotFound = errors.New("prodavnica nije pronađena")
	// ErrValidation se vraća za nevalidan ulaz.
	ErrValidation = errors.New("nevalidan ulaz")
)

// ShopResource je željeno stanje Shop CR-a koje Orchestrator primenjuje
// u klasteru. Namerno je nezavisno od GORM modela.
type ShopResource struct {
	Name          string // ime Shop CR-a u klasteru (npr. shop-<uuid>)
	Namespace     string
	DisplayName   string
	Availability  string
	DatabaseType  string
	WalletAddress string
	Image         string
}

// ResourceStatus je posmatrano stanje Shop CR-a pročitano iz klastera.
type ResourceStatus struct {
	Phase string // .status.phase: Pending|Provisioning|Ready|Failed (ili "")
}

// Orchestrator apstrahuje Kubernetes operacije nad Shop CR-ovima.
// Implementira ga internal/k8s.Client (in-cluster/kubeconfig), a u testovima
// i lokalno bez klastera koristi se no-op implementacija.
type Orchestrator interface {
	// Apply kreira ili ažurira Shop CR (idempotentno).
	Apply(ctx context.Context, r ShopResource) error
	// Status čita .status.phase datog Shop CR-a.
	Status(ctx context.Context, namespace, name string) (ResourceStatus, error)
	// Delete briše Shop CR; ne sme da greši ako CR ne postoji.
	Delete(ctx context.Context, namespace, name string) error
}

// Repository apstrahuje perzistenciju Shop meta-podataka.
type Repository interface {
	Create(ctx context.Context, shop *models.Shop) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Shop, error)
	GetByID(ctx context.Context, ownerID, id uuid.UUID) (*models.Shop, error)
	Update(ctx context.Context, shop *models.Shop) error
	Delete(ctx context.Context, ownerID, id uuid.UUID) error
}

// CreateInput je ulaz za kreiranje prodavnice (POST /shops).
type CreateInput struct {
	Name          string
	Availability  string
	WalletAddress string
	DatabaseType  string
	Image         string // opciono; ako je prazno koristi se podrazumevana slika
}

// UpdateInput je delimičan ulaz za PATCH /shops/{id}. Nil polja se ne menjaju.
type UpdateInput struct {
	Availability  *string
	WalletAddress *string
	DatabaseType  *string
}

// View je odgovor servisa: meta-podaci iz baze + status pročitan iz klastera.
type View struct {
	Shop   models.Shop
	Status string // pending|deploying|running|error
	URL    string
}

// Service je javni interfejs poslovne logike prodavnica (mock-uje se u testovima handler-a).
type Service interface {
	List(ctx context.Context, ownerID uuid.UUID) ([]View, error)
	Create(ctx context.Context, ownerID uuid.UUID, in CreateInput) (*View, error)
	Get(ctx context.Context, ownerID, id uuid.UUID) (*View, error)
	Update(ctx context.Context, ownerID, id uuid.UUID, in UpdateInput) (*View, error)
	Delete(ctx context.Context, ownerID, id uuid.UUID) error
	URL(ctx context.Context, ownerID, id uuid.UUID) (string, error)
}

// Options konfiguriše servis.
type Options struct {
	Namespace    string // namespace u koji idu Shop CR-ovi
	DefaultImage string // podrazumevana slika ako korisnik ne navede
	URLTemplate  string // npr. "http://%s.shophub.local" (%s = ime CR-a)
}

type service struct {
	repo  Repository
	orch  Orchestrator
	nsp   string
	image string
	tmpl  string
}

// NewService kreira servis prodavnica.
func NewService(repo Repository, orch Orchestrator, opts Options) Service {
	if opts.Namespace == "" {
		opts.Namespace = "default"
	}
	if opts.DefaultImage == "" {
		opts.DefaultImage = "ghcr.io/shophub-platform/shop:latest"
	}
	if opts.URLTemplate == "" {
		opts.URLTemplate = "http://%s.shophub.local"
	}
	return &service{repo: repo, orch: orch, nsp: opts.Namespace, image: opts.DefaultImage, tmpl: opts.URLTemplate}
}

// CRName vraća deterministično ime Shop CR-a za dati id prodavnice.
func CRName(id uuid.UUID) string { return "shop-" + id.String() }

// phaseToStatus mapira CRD phase na javni status iz OpenAPI specifikacije.
func phaseToStatus(phase string) string {
	switch phase {
	case "Ready":
		return "running"
	case "Provisioning":
		return "deploying"
	case "Failed":
		return "error"
	default: // Pending ili prazno
		return "pending"
	}
}

func validAvailability(v string) bool { return v == "standard" || v == "high" }
func validDatabase(v string) bool     { return v == "postgres" || v == "redis" }

func (s *service) url(shop models.Shop) string {
	return fmt.Sprintf(s.tmpl, CRName(shop.ID))
}

func (s *service) resourceFor(shop models.Shop) ShopResource {
	img := shop.Image
	if img == "" {
		img = s.image
	}
	return ShopResource{
		Name:          CRName(shop.ID),
		Namespace:     s.nsp,
		DisplayName:   shop.Name,
		Availability:  string(shop.Availability),
		DatabaseType:  string(shop.DatabaseType),
		WalletAddress: shop.WalletAddr,
		Image:         img,
	}
}

// viewFor sklapa View i čita aktuelni status iz klastera (best-effort).
func (s *service) viewFor(ctx context.Context, shop models.Shop) *View {
	v := &View{Shop: shop, Status: "pending", URL: s.url(shop)}
	if st, err := s.orch.Status(ctx, s.nsp, CRName(shop.ID)); err == nil {
		v.Status = phaseToStatus(st.Phase)
	}
	return v
}

func (s *service) List(ctx context.Context, ownerID uuid.UUID) ([]View, error) {
	shops, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(shops))
	for _, sh := range shops {
		views = append(views, *s.viewFor(ctx, sh))
	}
	return views, nil
}

func (s *service) Create(ctx context.Context, ownerID uuid.UUID, in CreateInput) (*View, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, fmt.Errorf("%w: ime je obavezno", ErrValidation)
	}
	if in.Availability == "" {
		in.Availability = "standard"
	}
	if in.DatabaseType == "" {
		in.DatabaseType = "postgres"
	}
	if !validAvailability(in.Availability) {
		return nil, fmt.Errorf("%w: availability mora biti standard ili high", ErrValidation)
	}
	if !validDatabase(in.DatabaseType) {
		return nil, fmt.Errorf("%w: databaseType mora biti postgres ili redis", ErrValidation)
	}
	if strings.TrimSpace(in.WalletAddress) == "" {
		return nil, fmt.Errorf("%w: walletAddress je obavezan", ErrValidation)
	}
	image := strings.TrimSpace(in.Image)
	if image == "" {
		image = s.image
	}

	shop := &models.Shop{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         in.Name,
		Availability: models.Availability(in.Availability),
		WalletAddr:   in.WalletAddress,
		DatabaseType: models.DatabaseType(in.DatabaseType),
		Image:        image,
	}

	// Prvo upiši CR; ako uspe, perzistuj meta-podatke. Tako baza ne sadrži
	// prodavnice za koje provisioning nikad nije ni započeo.
	if err := s.orch.Apply(ctx, s.resourceFor(*shop)); err != nil {
		return nil, fmt.Errorf("kreiranje Shop CR-a: %w", err)
	}
	if err := s.repo.Create(ctx, shop); err != nil {
		// Pokušaj da očistiš CR da ne ostane siroče.
		_ = s.orch.Delete(ctx, s.nsp, CRName(shop.ID))
		return nil, err
	}
	metrics.ShopsCreatedTotal.Inc()
	return s.viewFor(ctx, *shop), nil
}

func (s *service) Get(ctx context.Context, ownerID, id uuid.UUID) (*View, error) {
	shop, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	return s.viewFor(ctx, *shop), nil
}

func (s *service) Update(ctx context.Context, ownerID, id uuid.UUID, in UpdateInput) (*View, error) {
	shop, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if in.Availability != nil {
		if !validAvailability(*in.Availability) {
			return nil, fmt.Errorf("%w: availability mora biti standard ili high", ErrValidation)
		}
		shop.Availability = models.Availability(*in.Availability)
	}
	if in.DatabaseType != nil {
		if !validDatabase(*in.DatabaseType) {
			return nil, fmt.Errorf("%w: databaseType mora biti postgres ili redis", ErrValidation)
		}
		shop.DatabaseType = models.DatabaseType(*in.DatabaseType)
	}
	if in.WalletAddress != nil {
		if strings.TrimSpace(*in.WalletAddress) == "" {
			return nil, fmt.Errorf("%w: walletAddress ne sme biti prazan", ErrValidation)
		}
		shop.WalletAddr = *in.WalletAddress
	}

	if err := s.orch.Apply(ctx, s.resourceFor(*shop)); err != nil {
		return nil, fmt.Errorf("ažuriranje Shop CR-a: %w", err)
	}
	if err := s.repo.Update(ctx, shop); err != nil {
		return nil, err
	}
	return s.viewFor(ctx, *shop), nil
}

func (s *service) Delete(ctx context.Context, ownerID, id uuid.UUID) error {
	// Potvrdi vlasništvo pre brisanja.
	shop, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return err
	}
	if err := s.orch.Delete(ctx, s.nsp, CRName(shop.ID)); err != nil {
		return fmt.Errorf("brisanje Shop CR-a: %w", err)
	}
	return s.repo.Delete(ctx, ownerID, id)
}

func (s *service) URL(ctx context.Context, ownerID, id uuid.UUID) (string, error) {
	shop, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return "", err
	}
	return s.url(*shop), nil
}
