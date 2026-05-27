package helper

import "testing"

func TestPruneEmptyFields(t *testing.T) {
	t.Parallel()

	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "demo",
			"namespace": "default",
			"labels":    map[string]interface{}{},
		},
		"spec": map[string]interface{}{
			"hosts": []interface{}{},
			"ports": []interface{}{
				map[string]interface{}{
					"name":     "http",
					"number":   float64(80),
					"protocol": "",
				},
			},
		},
	}

	PruneEmptyFields(obj)

	if _, ok := obj["metadata"].(map[string]interface{})["labels"]; ok {
		t.Fatal("expected empty labels map to be pruned")
	}
	if _, ok := obj["spec"].(map[string]interface{})["hosts"]; ok {
		t.Fatal("expected empty hosts slice to be pruned")
	}
	ports := obj["spec"].(map[string]interface{})["ports"].([]interface{})
	port := ports[0].(map[string]interface{})
	if _, ok := port["protocol"]; ok {
		t.Fatal("expected empty protocol to be pruned")
	}
}
