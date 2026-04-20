package cmd

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/aohoyd/aku/internal/msgs"
	"github.com/aohoyd/aku/internal/plugin"
	"github.com/aohoyd/aku/internal/render"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// mockPlugin implements plugin.ResourcePlugin for testing.
type mockPlugin struct {
	name      string
	shortName string
	gvr       schema.GroupVersionResource
}

func (m *mockPlugin) Name() string                              { return m.name }
func (m *mockPlugin) ShortName() string                         { return m.shortName }
func (m *mockPlugin) GVR() schema.GroupVersionResource          { return m.gvr }
func (m *mockPlugin) IsClusterScoped() bool                     { return false }
func (m *mockPlugin) Columns() []plugin.Column                  { return nil }
func (m *mockPlugin) Row(_ *unstructured.Unstructured) []string { return nil }
func (m *mockPlugin) YAML(_ *unstructured.Unstructured) (render.Content, error) {
	return render.Content{}, nil
}
func (m *mockPlugin) Describe(_ context.Context, _ *unstructured.Unstructured) (render.Content, error) {
	return render.Content{}, nil
}

// registerTestPlugins resets the plugin registry and registers a standard set of
// mock plugins used across multiple tests. Call this at the start of each test
// that relies on plugin lookups to ensure a clean, predictable state.
func registerTestPlugins() {
	plugin.Reset()
	plugin.Register(&mockPlugin{
		name:      "pods",
		shortName: "po",
		gvr:       schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
	})
	plugin.Register(&mockPlugin{
		name:      "deployments",
		shortName: "deploy",
		gvr:       schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
	})
	plugin.Register(&mockPlugin{
		name:      "secrets",
		shortName: "sec",
		gvr:       schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"},
	})
	plugin.Register(&mockPlugin{
		name:      "services",
		shortName: "svc",
		gvr:       schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"},
	})
}

func TestParseResourceSpecs_CommaSplitting(t *testing.T) {
	registerTestPlugins()

	specs, err := parseResourceSpecs([]string{"po", "deploy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(specs))
	}
	if specs[0].Plugin.Name() != "pods" {
		t.Errorf("expected first spec plugin name 'pods', got %q", specs[0].Plugin.Name())
	}
	if specs[1].Plugin.Name() != "deployments" {
		t.Errorf("expected second spec plugin name 'deployments', got %q", specs[1].Plugin.Name())
	}
}

func TestParseResourceSpecs_NamespacePrefix(t *testing.T) {
	registerTestPlugins()

	specs, err := parseResourceSpecs([]string{"kube-system/sec"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].Namespace != "kube-system" {
		t.Errorf("expected namespace 'kube-system', got %q", specs[0].Namespace)
	}
	if specs[0].Plugin.Name() != "secrets" {
		t.Errorf("expected plugin name 'secrets', got %q", s