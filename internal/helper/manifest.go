package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

// GVRFromManifest derives GroupVersionResource from the manifest apiVersion.
// Falls back to defaultGVR when apiVersion is missing or unparseable.
func GVRFromManifest(defaultGVR k8sschema.GroupVersionResource, manifest map[string]interface{}) k8sschema.GroupVersionResource {
	apiVersion, ok := manifest["apiVersion"].(string)
	if !ok || apiVersion == "" {
		return defaultGVR
	}
	parts := strings.SplitN(apiVersion, "/", 2)
	if len(parts) != 2 {
		return defaultGVR
	}
	return k8sschema.GroupVersionResource{
		Group:    parts[0],
		Version:  parts[1],
		Resource: defaultGVR.Resource,
	}
}

// ParseManifestJSON unmarshals a manifest JSON string into a map.
func ParseManifestJSON(raw string) (map[string]interface{}, error) {
	var manifest map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}
	return manifest, nil
}

// ValidateManifest checks required Kubernetes fields before an API call.
func ValidateManifest(manifest map[string]interface{}, expectedKind string, requireNamespace bool) error {
	apiVersion, ok := manifest["apiVersion"].(string)
	if !ok || apiVersion == "" {
		return fmt.Errorf("manifest must include apiVersion")
	}
	kind, ok := manifest["kind"].(string)
	if !ok || kind == "" {
		return fmt.Errorf("manifest must include kind")
	}
	if kind != expectedKind {
		return fmt.Errorf("manifest kind %q does not match expected %q", kind, expectedKind)
	}
	meta, ok := manifest["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("manifest must include metadata")
	}
	name, ok := meta["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("manifest metadata must include name")
	}
	if requireNamespace {
		ns, ok := meta["namespace"].(string)
		if !ok || ns == "" {
			return fmt.Errorf("manifest metadata must include namespace")
		}
	}
	return nil
}

// ManifestJSONFromObject normalizes a cluster object for Terraform state (strip, prune, redact, canonical JSON).
func ManifestJSONFromObject(obj map[string]interface{}) (string, error) {
	return NormalizeManifestForState(obj)
}

// MarshalCanonicalJSON encodes a value with sorted map keys for stable comparisons.
func MarshalCanonicalJSON(v interface{}) ([]byte, error) {
	switch x := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')
			valBytes, err := MarshalCanonicalJSON(x[k])
			if err != nil {
				return nil, err
			}
			buf.Write(valBytes)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case []interface{}:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, elem := range x {
			if i > 0 {
				buf.WriteByte(',')
			}
			valBytes, err := MarshalCanonicalJSON(elem)
			if err != nil {
				return nil, err
			}
			buf.Write(valBytes)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	default:
		return json.Marshal(v)
	}
}
