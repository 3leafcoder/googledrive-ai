package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The release workflow is config, not code, so it can't be exercised locally.
// This test parses .github/workflows/release.yml and pins the wiring the
// ci-publishes-release-assets assertion depends on: the version-tag trigger,
// the test-before-build ordering, the shared cross-compile step with the
// tag-derived version, and a maintained release action that uploads the assets.
type workflow struct {
	On   map[string]any `yaml:"on"`
	Jobs map[string]struct {
		Steps []struct {
			Name string         `yaml:"name"`
			Run  string         `yaml:"run"`
			Uses string         `yaml:"uses"`
			With map[string]any `yaml:"with"`
		} `yaml:"steps"`
	} `yaml:"jobs"`
}

func loadWorkflow(t *testing.T) workflow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("parse release workflow: %v", err)
	}
	return wf
}

func TestReleaseWorkflowTriggers(t *testing.T) {
	wf := loadWorkflow(t)

	if _, ok := wf.On["workflow_dispatch"]; !ok {
		t.Error("missing workflow_dispatch trigger for on-demand releases")
	}

	push, _ := wf.On["push"].(map[string]any)
	tags, _ := push["tags"].([]any)
	found := false
	for _, tg := range tags {
		if s, _ := tg.(string); s == "v*" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a v* tag trigger, got tags %v", tags)
	}
}

func TestReleaseWorkflowStepOrdering(t *testing.T) {
	wf := loadWorkflow(t)
	job, ok := wf.Jobs["release"]
	if !ok {
		t.Fatal("no release job defined")
	}

	testIdx, buildIdx, uploadIdx := -1, -1, -1
	var versionExpr string
	for i, s := range job.Steps {
		switch {
		case strings.Contains(s.Run, "make test"):
			testIdx = i
		case strings.Contains(s.Run, "make build-all"):
			buildIdx = i
			versionExpr = s.Run
		case strings.HasPrefix(s.Uses, "softprops/action-gh-release"):
			uploadIdx = i
			// The action must actually attach the built assets.
			if files, _ := s.With["files"].(string); !strings.Contains(files, "dist") {
				t.Errorf("release action uploads %q, want the dist assets", files)
			}
		}
	}

	if testIdx < 0 {
		t.Fatal("no `make test` step: tests must gate the release")
	}
	if buildIdx < 0 {
		t.Fatal("no `make build-all` step: release must use the shared cross-compile")
	}
	if uploadIdx < 0 {
		t.Fatal("no softprops/action-gh-release step to publish the assets")
	}
	// A test failure must abort before building or uploading.
	if !(testIdx < buildIdx && buildIdx < uploadIdx) {
		t.Errorf("steps out of order: test=%d build=%d upload=%d (want test<build<upload)", testIdx, buildIdx, uploadIdx)
	}
	// The version passed to the build must come from the tag, not be hardcoded.
	if !strings.Contains(versionExpr, "VERSION=") || !strings.Contains(versionExpr, "version") {
		t.Errorf("build step %q does not pass a tag-derived VERSION", versionExpr)
	}
}
