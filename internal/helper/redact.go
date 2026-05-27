package helper

import (
	"strings"
)

// RedactedPlaceholder replaces sensitive values in state so secrets are not persisted.
const RedactedPlaceholder = "(redacted)"

// sensitiveFieldNames lists JSON object keys whose values are redacted in Terraform state.
var sensitiveFieldNames = map[string]struct{}{
	"token":          {},
	"password":       {},
	"secret":         {},
	"privatekey":     {},
	"private_key":    {},
	"clientsecret":   {},
	"client_secret":  {},
	"credential":     {},
	"credentials":    {},
	"cacert":         {},
	"ca_cert":        {},
	"cacertbase64":   {},
	"ca_cert_base64": {},
	"cakey":          {},
	"ca_key":         {},
	"cert":           {},
	"key":            {},
	"stringdata":     {},
	"rootca":         {},
	"root_ca":        {},
}

// RedactSensitiveFields deep-copies and replaces values at sensitive keys and annotation paths.
func RedactSensitiveFields(obj map[string]interface{}) {
	redactValue(obj)
}

func redactValue(v interface{}) {
	switch x := v.(type) {
	case map[string]interface{}:
		for k, val := range x {
			if isSensitiveField(k) {
				x[k] = RedactedPlaceholder
				continue
			}
			if k == "annotations" {
				if ann, ok := val.(map[string]interface{}); ok {
					redactAnnotations(ann)
				}
				continue
			}
			redactValue(val)
		}
	case []interface{}:
		for _, elem := range x {
			redactValue(elem)
		}
	}
}

func redactAnnotations(ann map[string]interface{}) {
	for k := range ann {
		if isSensitiveAnnotationKey(k) {
			ann[k] = RedactedPlaceholder
		}
	}
}

func isSensitiveField(name string) bool {
	_, ok := sensitiveFieldNames[strings.ToLower(name)]
	return ok
}

func isSensitiveAnnotationKey(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "token") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "credential")
}

// ManifestsEqualIgnoringRedacted reports whether two manifests match, treating
// redacted placeholders as equal to any value at the same path.
func ManifestsEqualIgnoringRedacted(aJSON, bJSON string) (bool, error) {
	a, err := ParseManifestJSON(aJSON)
	if err != nil {
		return false, err
	}
	b, err := ParseManifestJSON(bJSON)
	if err != nil {
		return false, err
	}
	return manifestMapsEqual(a, b), nil
}

func manifestMapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, aVal := range a {
		bVal, ok := b[k]
		if !ok || !valuesEqualIgnoringRedacted(aVal, bVal) {
			return false
		}
	}
	return true
}

func valuesEqualIgnoringRedacted(a, b interface{}) bool {
	if isRedactedValue(a) || isRedactedValue(b) {
		return true
	}
	switch aTyped := a.(type) {
	case map[string]interface{}:
		bTyped, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return manifestMapsEqual(aTyped, bTyped)
	case []interface{}:
		bTyped, ok := b.([]interface{})
		if !ok || len(aTyped) != len(bTyped) {
			return false
		}
		for i := range aTyped {
			if !valuesEqualIgnoringRedacted(aTyped[i], bTyped[i]) {
				return false
			}
		}
		return true
	default:
		return deepEqualScalars(a, b)
	}
}

func isRedactedValue(v interface{}) bool {
	s, ok := v.(string)
	return ok && s == RedactedPlaceholder
}

func deepEqualScalars(a, b interface{}) bool {
	switch aVal := a.(type) {
	case string:
		bVal, ok := b.(string)
		return ok && aVal == bVal
	case float64:
		bVal, ok := b.(float64)
		return ok && aVal == bVal
	case bool:
		bVal, ok := b.(bool)
		return ok && aVal == bVal
	case nil:
		return b == nil
	default:
		return false
	}
}

// ManifestContainsRedacted returns true if the JSON contains a redacted placeholder.
func ManifestContainsRedacted(manifestJSON string) bool {
	return strings.Contains(manifestJSON, RedactedPlaceholder)
}
