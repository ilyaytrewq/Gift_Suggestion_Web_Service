package postgres

import (
	"encoding/json"
	"testing"
)

func TestMarshalScopesNilIsJSONArray(t *testing.T) {
	t.Parallel()

	raw, err := marshalScopes(nil)
	if err != nil {
		t.Fatalf("marshalScopes(nil) error = %v", err)
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	list, ok := decoded.([]any)
	if !ok {
		t.Fatalf("expected JSON array, got %T (%s)", decoded, string(raw))
	}
	if len(list) != 0 {
		t.Fatalf("expected empty array, got %v", list)
	}
}

func TestMarshalScopesPreservesValues(t *testing.T) {
	t.Parallel()

	raw, err := marshalScopes([]string{"groups", "offline"})
	if err != nil {
		t.Fatalf("marshalScopes() error = %v", err)
	}

	var decoded []string
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded) != 2 || decoded[0] != "groups" || decoded[1] != "offline" {
		t.Fatalf("unexpected scopes: %v", decoded)
	}
}
