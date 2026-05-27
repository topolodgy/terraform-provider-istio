package helper

import "testing"

func TestRedactSensitiveFields(t *testing.T) {
	t.Parallel()

	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"jwtRules": []interface{}{
				map[string]interface{}{
					"issuer": "https://example.com",
					"token":  "secret-key-material",
				},
			},
		},
		"metadata": map[string]interface{}{
			"name": "demo",
			"annotations": map[string]interface{}{
				"proxy.istio.io/config":           "normal",
				"vault.hashicorp.com/agent-token": "tok123",
			},
		},
	}

	RedactSensitiveFields(obj)

	jwt := obj["spec"].(map[string]interface{})["jwtRules"].([]interface{})[0].(map[string]interface{})
	if jwt["token"] != RedactedPlaceholder {
		t.Fatalf("expected token field redacted, got %#v", jwt["token"])
	}

	ann := obj["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	if ann["proxy.istio.io/config"] != "normal" {
		t.Fatal("expected non-sensitive annotation to remain")
	}
	if ann["vault.hashicorp.com/agent-token"] != RedactedPlaceholder {
		t.Fatalf("expected token annotation redacted, got %#v", ann["vault.hashicorp.com/agent-token"])
	}
}

func TestManifestsEqualIgnoringRedacted(t *testing.T) {
	t.Parallel()

	state := `{"metadata":{"name":"x"},"spec":{"token":"(redacted)"}}`
	plan := `{"metadata":{"name":"x"},"spec":{"token":"real-secret"}}`

	equal, err := ManifestsEqualIgnoringRedacted(plan, state)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("expected plan and redacted state to be equal")
	}
}

func TestManifestsEqualIgnoringRedacted_DifferentSpec(t *testing.T) {
	t.Parallel()

	state := `{"metadata":{"name":"x"},"spec":{"hosts":["a"]}}`
	plan := `{"metadata":{"name":"x"},"spec":{"hosts":["b"]}}`

	equal, err := ManifestsEqualIgnoringRedacted(plan, state)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("expected different specs to be unequal")
	}
}
