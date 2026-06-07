package shops

import "context"

// NoopOrchestrator je bezopasna implementacija Orchestrator-a koja ne radi
// ništa. Koristi se lokalno kada Kubernetes klaster nije dostupan, kako bi
// REST API i dalje radio (CR-ovi se ne kreiraju). Status uvek "" (-> pending).
type NoopOrchestrator struct{}

// NewNoopOrchestrator vraća no-op orkestrator.
func NewNoopOrchestrator() Orchestrator { return NoopOrchestrator{} }

func (NoopOrchestrator) Apply(context.Context, ShopResource) error { return nil }

func (NoopOrchestrator) Status(context.Context, string, string) (ResourceStatus, error) {
	return ResourceStatus{}, nil
}

func (NoopOrchestrator) Delete(context.Context, string, string) error { return nil }
