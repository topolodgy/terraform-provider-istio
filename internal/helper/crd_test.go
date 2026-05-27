package helper

import (
	"strings"
	"testing"

	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

func TestStripServerFields(t *testing.T) {
	t.Parallel()

	obj := map[string]interface{}{
		"status": map[string]interface{}{"phase": "Active"},
		"metadata": map[string]interface{}{
			"name":              "demo",
			"namespace":         "default",
			"resourceVersion":   "123",
			"uid":               "abc",
			"creationTimestamp": "2024-01-01T00:00:00Z",
			"generation":        int64(1),
			"managedFields":     []interface{}{},
			"annotations": map[string]interface{}{
				"kubectl.kubernetes.io/last-applied-configuration": "{}",
				"custom": "keep",
			},
		},
		"spec": map[string]interface{}{"hosts": []interface{}{"example.com"}},
	}

	StripServerFields(obj)

	if _, ok := obj["status"]; ok {
		t.Fatal("expected status to be removed")
	}
	meta := obj["metadata"].(map[string]interface{})
	for _, key := range []string{"resourceVersion", "uid", "creationTimestamp", "generation", "managedFields"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("expected metadata.%s to be removed", key)
		}
	}
	ann := meta["annotations"].(map[string]interface{})
	if _, ok := ann["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		t.Fatal("expected last-applied-configuration annotation to be removed")
	}
	if ann["custom"] != "keep" {
		t.Fatalf("expected custom annotation to remain, got %#v", ann["custom"])
	}
}

func TestStripServerFields_EmptyAnnotationsRemoved(t *testing.T) {
	t.Parallel()

	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "demo",
			"annotations": map[string]interface{}{
				"kubectl.kubernetes.io/last-applied-configuration": "{}",
			},
		},
	}
	StripServerFields(obj)
	meta := obj["metadata"].(map[string]interface{})
	if _, ok := meta["annotations"]; ok {
		t.Fatal("expected empty annotations map to be removed")
	}
}

func TestParseId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       string
		wantNS   string
		wantName string
	}{
		{id: "default/my-resource", wantNS: "default", wantName: "my-resource"},
		{id: "cluster-scoped", wantNS: "", wantName: "cluster-scoped"},
		{id: "ns/with/slash", wantNS: "ns/with", wantName: "slash"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.id, func(t *testing.T) {
			t.Parallel()
			ns, name := ParseId(tc.id)
			if ns != tc.wantNS || name != tc.wantName {
				t.Fatalf("ParseId(%q) = (%q, %q), want (%q, %q)", tc.id, ns, name, tc.wantNS, tc.wantName)
			}
		})
	}
}

func TestGVRFromManifest(t *testing.T) {
	t.Parallel()

	defaultGVR := k8sschema.GroupVersionResource{
		Group: "networking.istio.io", Version: "v1", Resource: "serviceentries",
	}

	tests := []struct {
		name     string
		manifest map[string]interface{}
		want     k8sschema.GroupVersionResource
	}{
		{
			name:     "uses manifest apiVersion",
			manifest: map[string]interface{}{"apiVersion": "networking.istio.io/v1beta1"},
			want:     k8sschema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "serviceentries"},
		},
		{
			name:     "falls back when apiVersion missing",
			manifest: map[string]interface{}{"kind": "ServiceEntry"},
			want:     defaultGVR,
		},
		{
			name:     "falls back when apiVersion invalid",
			manifest: map[string]interface{}{"apiVersion": "invalid"},
			want:     defaultGVR,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := GVRFromManifest(defaultGVR, tc.manifest)
			if got != tc.want {
				t.Fatalf("GVRFromManifest() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	valid := map[string]interface{}{
		"apiVersion": "networking.istio.io/v1",
		"kind":       "ServiceEntry",
		"metadata": map[string]interface{}{
			"name":      "demo",
			"namespace": "default",
		},
	}

	if err := ValidateManifest(valid, "ServiceEntry", true); err != nil {
		t.Fatalf("expected valid manifest, got %v", err)
	}

	tests := []struct {
		name     string
		manifest map[string]interface{}
		wantErr  string
	}{
		{
			name:     "missing apiVersion",
			manifest: map[string]interface{}{"kind": "ServiceEntry", "metadata": map[string]interface{}{"name": "x", "namespace": "default"}},
			wantErr:  "apiVersion",
		},
		{
			name:     "wrong kind",
			manifest: map[string]interface{}{"apiVersion": "networking.istio.io/v1", "kind": "Gateway", "metadata": map[string]interface{}{"name": "x", "namespace": "default"}},
			wantErr:  "does not match",
		},
		{
			name:     "missing namespace",
			manifest: map[string]interface{}{"apiVersion": "networking.istio.io/v1", "kind": "ServiceEntry", "metadata": map[string]interface{}{"name": "x"}},
			wantErr:  "namespace",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateManifest(tc.manifest, "ServiceEntry", true)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q should contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestMarshalCanonicalJSON_StableKeyOrder(t *testing.T) {
	t.Parallel()

	obj := map[string]interface{}{
		"z": 1,
		"a": map[string]interface{}{"y": 2, "b": 3},
	}
	first, err := MarshalCanonicalJSON(obj)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalCanonicalJSON(map[string]interface{}{
		"a": map[string]interface{}{"b": 3, "y": 2},
		"z": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("canonical JSON differs:\n%s\n%s", first, second)
	}
}
