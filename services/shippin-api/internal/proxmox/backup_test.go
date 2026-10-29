package proxmox

import "testing"

func TestWorkspaceExcludesSkipCaches(t *testing.T) {
	want := map[string]bool{
		"cache": true, "scratch/tmp": true, "**/node_modules": true, "**/target": true,
	}
	for _, ex := range WorkspaceExcludes {
		delete(want, ex)
	}
	if len(want) != 0 {
		t.Fatalf("missing excludes: %v", want)
	}
}
