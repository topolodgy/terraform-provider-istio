package helper

// NormalizeManifestForState prepares a cluster object for Terraform state storage.
func NormalizeManifestForState(obj map[string]interface{}) (string, error) {
	StripServerFields(obj)
	PruneEmptyFields(obj)
	RedactSensitiveFields(obj)
	raw, err := MarshalCanonicalJSON(obj)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// PruneEmptyFields removes empty strings, nulls, empty maps, and empty slices recursively.
func PruneEmptyFields(obj map[string]interface{}) {
	for k, v := range obj {
		switch x := v.(type) {
		case map[string]interface{}:
			PruneEmptyFields(x)
			if len(x) == 0 {
				delete(obj, k)
			}
		case []interface{}:
			pruned := pruneSlice(x)
			if len(pruned) == 0 {
				delete(obj, k)
			} else {
				obj[k] = pruned
			}
		case string:
			if x == "" {
				delete(obj, k)
			}
		case nil:
			delete(obj, k)
		}
	}
}

func pruneSlice(items []interface{}) []interface{} {
	out := make([]interface{}, 0, len(items))
	for _, item := range items {
		switch x := item.(type) {
		case map[string]interface{}:
			PruneEmptyFields(x)
			if len(x) > 0 {
				out = append(out, x)
			}
		case []interface{}:
			nested := pruneSlice(x)
			if len(nested) > 0 {
				out = append(out, nested)
			}
		case string:
			if x != "" {
				out = append(out, x)
			}
		case nil:
			continue
		default:
			out = append(out, item)
		}
	}
	return out
}
