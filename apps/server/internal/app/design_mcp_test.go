package app

import "testing"

func TestDesignIDsAreUnique(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for range 1000 {
		id := designID()
		if len(id) != 32 {
			t.Fatalf("designID() length = %d, want 32", len(id))
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("designID() returned duplicate %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestDesignMCPToolContract(t *testing.T) {
	tools := designMCPTools()
	want := map[string]bool{"list_projects": false, "register_project": false, "create_design_session": false, "read_design_context": false, "submit_proposal": false, "read_reviews": false}
	for _, tool := range tools {
		name, ok := tool["name"].(string)
		if !ok {
			t.Fatalf("tool has no string name: %#v", tool)
		}
		if _, exists := want[name]; !exists {
			t.Errorf("unexpected tool %q", name)
			continue
		}
		want[name] = true
		if _, ok := tool["inputSchema"].(map[string]any); !ok {
			t.Errorf("tool %q has no input schema", name)
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing tool %q", name)
		}
	}
}
