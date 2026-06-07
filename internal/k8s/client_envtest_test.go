//go:build integration

// Integracioni test K8s integracije preko envtest (in-memory API server iz
// controller-runtime) (FZ 2.3).
//
// Pokretanje:
//   go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
//   export KUBEBUILDER_ASSETS=$(setup-envtest use -p path 1.30.0)
//   go test -tags=integration ./internal/k8s/ -run Envtest -v
package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/shophub-platform/shophub/internal/shops"
)

func TestEnvtestShopCRLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	env := &envtest.Environment{
		CRDDirectoryPaths:     []string{"../../test/crds"},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err := env.Start()
	require.NoError(t, err, "envtest start (da li je KUBEBUILDER_ASSETS postavljen?)")
	defer func() { _ = env.Stop() }()

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)
	client := NewClientFromDynamic(dyn)

	res := shops.ShopResource{
		Name:          "shop-envtest",
		Namespace:     "default",
		DisplayName:   "Envtest Radnja",
		Availability:  "high",
		DatabaseType:  "postgres",
		WalletAddress: "0xdeadbeef",
		Image:         "ghcr.io/shophub-platform/shop:latest",
	}

	// Apply (create)
	require.NoError(t, client.Apply(ctx, res))

	// CR mora biti vidljiv preko dinamičkog klijenta (ekvivalent kubectl get shops)
	got, err := dyn.Resource(shopGVR).Namespace("default").Get(ctx, res.Name, metav1.GetOptions{})
	require.NoError(t, err)
	name, _, _ := unstructured.NestedString(got.Object, "spec", "name")
	require.Equal(t, "Envtest Radnja", name)
	avail, _, _ := unstructured.NestedString(got.Object, "spec", "availability")
	require.Equal(t, "high", avail)

	// Apply (update) — promena dostupnosti je idempotentna
	res.Availability = "standard"
	require.NoError(t, client.Apply(ctx, res))
	got, _ = dyn.Resource(shopGVR).Namespace("default").Get(ctx, res.Name, metav1.GetOptions{})
	avail, _, _ = unstructured.NestedString(got.Object, "spec", "availability")
	require.Equal(t, "standard", avail)

	// Status (bez kontrolera phase je prazan -> nije greška)
	st, err := client.Status(ctx, "default", res.Name)
	require.NoError(t, err)
	require.Equal(t, "", st.Phase)

	// Delete je idempotentan
	require.NoError(t, client.Delete(ctx, "default", res.Name))
	require.NoError(t, client.Delete(ctx, "default", res.Name)) // ponovni delete ne sme da greši
}
