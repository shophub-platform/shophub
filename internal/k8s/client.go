// Package k8s implementira shops.Orchestrator preko client-go dinamičkog
// klijenta. Kreira/ažurira/briše Shop CR-ove (grupa shop.shophub.io) i čita
// njihov status. Koristi in-cluster konfiguraciju u podu, a kubeconfig lokalno.
package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/shophub-platform/shophub/internal/shops"
)

const (
	crGroup   = "shop.shophub.io"
	crVersion = "v1alpha1"
	crKind    = "Shop"

	walletAddrAnnotation = "shop.shophub.io/wallet-address"
	managedByLabel       = "app.kubernetes.io/managed-by"
	managedByValue       = "shophub"
)

// shopGVR je GroupVersionResource Shop CRD-a (resurs je u množini, malim slovima).
var shopGVR = schema.GroupVersionResource{Group: crGroup, Version: crVersion, Resource: "shops"}

// Client je client-go implementacija shops.Orchestrator-a.
type Client struct {
	dyn dynamic.Interface
}

// statički dokaz da Client zadovoljava interfejs.
var _ shops.Orchestrator = (*Client)(nil)

// NewClient pravi klijent. Prvo pokušava in-cluster config (kada radi u podu),
// pa pada na kubeconfig (eksplicitna putanja, KUBECONFIG env ili ~/.kube/config).
func NewClient(kubeconfigPath string) (*Client, error) {
	cfg, err := loadConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("dinamički klijent: %w", err)
	}
	return &Client{dyn: dyn}, nil
}

// NewClientFromDynamic je korisno za testove (envtest) gde već imamo config.
func NewClientFromDynamic(dyn dynamic.Interface) *Client { return &Client{dyn: dyn} }

func loadConfig(kubeconfigPath string) (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	if kubeconfigPath == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeconfigPath = env
		} else if home, _ := os.UserHomeDir(); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig (%q): %w", kubeconfigPath, err)
	}
	return cfg, nil
}

func desiredObject(r shops.ShopResource) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: crGroup, Version: crVersion, Kind: crKind})
	obj.SetName(r.Name)
	obj.SetNamespace(r.Namespace)
	obj.SetLabels(map[string]string{managedByLabel: managedByValue})
	obj.SetAnnotations(map[string]string{walletAddrAnnotation: r.WalletAddress})
	spec := map[string]interface{}{
		"name":         r.DisplayName,
		"availability": r.Availability,
		"databaseType": r.DatabaseType,
		"image":        r.Image,
		"walletRef":    map[string]interface{}{"name": r.Name + "-wallet"},
	}
	_ = unstructured.SetNestedMap(obj.Object, spec, "spec")
	return obj
}

// Apply kreira CR ako ne postoji, u suprotnom ažurira njegov spec/anotacije.
func (c *Client) Apply(ctx context.Context, r shops.ShopResource) error {
	ri := c.dyn.Resource(shopGVR).Namespace(r.Namespace)
	desired := desiredObject(r)

	existing, err := ri.Get(ctx, r.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err := ri.Create(ctx, desired, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("kreiranje Shop CR-a %q: %w", r.Name, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("čitanje Shop CR-a %q: %w", r.Name, err)
	}

	// Zadrži resourceVersion radi optimističkog zaključavanja, prepiši spec/meta.
	desired.SetResourceVersion(existing.GetResourceVersion())
	if _, err := ri.Update(ctx, desired, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("ažuriranje Shop CR-a %q: %w", r.Name, err)
	}
	return nil
}

// Status čita .status.phase Shop CR-a.
func (c *Client) Status(ctx context.Context, namespace, name string) (shops.ResourceStatus, error) {
	obj, err := c.dyn.Resource(shopGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return shops.ResourceStatus{}, err
	}
	phase, _, _ := unstructured.NestedString(obj.Object, "status", "phase")
	return shops.ResourceStatus{Phase: phase}, nil
}

// Delete briše Shop CR; nepostojeći CR se tretira kao uspeh (idempotentno).
func (c *Client) Delete(ctx context.Context, namespace, name string) error {
	err := c.dyn.Resource(shopGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("brisanje Shop CR-a %q: %w", name, err)
	}
	return nil
}
