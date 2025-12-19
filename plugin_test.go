package main

import (
	"context"
	"testing"

	"github.com/relicta-tech/relicta-plugin-sdk/plugin"
)

func TestGetInfo(t *testing.T) {
	p := &GCRPlugin{}
	info := p.GetInfo()

	if info.Name != "gcr" {
		t.Errorf("expected name 'gcr', got '%s'", info.Name)
	}

	if info.Description == "" {
		t.Error("expected non-empty description")
	}

	if len(info.Hooks) == 0 {
		t.Error("expected at least one hook")
	}

	hasPostPublish := false
	for _, hook := range info.Hooks {
		if hook == plugin.HookPostPublish {
			hasPostPublish = true
			break
		}
	}
	if !hasPostPublish {
		t.Error("expected PostPublish hook")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		config     map[string]any
		wantErrors int
	}{
		{
			name:       "empty config",
			config:     map[string]any{},
			wantErrors: 4, // project, image, source_image, repository (AR default)
		},
		{
			name: "missing project",
			config: map[string]any{
				"image":        "my-app",
				"source_image": "myapp:latest",
				"repository":   "my-repo",
			},
			wantErrors: 1,
		},
		{
			name: "missing image",
			config: map[string]any{
				"project":      "my-project",
				"source_image": "myapp:latest",
				"repository":   "my-repo",
			},
			wantErrors: 1,
		},
		{
			name: "missing source image",
			config: map[string]any{
				"project":    "my-project",
				"image":      "my-app",
				"repository": "my-repo",
			},
			wantErrors: 1,
		},
		{
			name: "artifact registry without repository",
			config: map[string]any{
				"project":          "my-project",
				"image":            "my-app",
				"source_image":     "myapp:latest",
				"artifact_registry": true,
			},
			wantErrors: 1,
		},
		{
			name: "legacy GCR without repository",
			config: map[string]any{
				"project":          "my-project",
				"image":            "my-app",
				"source_image":     "myapp:latest",
				"artifact_registry": false,
			},
			wantErrors: 0, // Legacy GCR doesn't need repository
		},
		{
			name: "invalid auth method",
			config: map[string]any{
				"project":      "my-project",
				"image":        "my-app",
				"source_image": "myapp:latest",
				"repository":   "my-repo",
				"auth": map[string]any{
					"method": "invalid",
				},
			},
			wantErrors: 1,
		},
		{
			name: "service account without key",
			config: map[string]any{
				"project":      "my-project",
				"image":        "my-app",
				"source_image": "myapp:latest",
				"repository":   "my-repo",
				"auth": map[string]any{
					"method": "service_account",
				},
			},
			wantErrors: 1,
		},
		{
			name: "valid config with gcloud auth",
			config: map[string]any{
				"project":      "my-project",
				"image":        "my-app",
				"source_image": "myapp:latest",
				"repository":   "my-repo",
				"auth": map[string]any{
					"method": "gcloud",
				},
			},
			wantErrors: 0,
		},
		{
			name: "valid config with service account key file",
			config: map[string]any{
				"project":      "my-project",
				"image":        "my-app",
				"source_image": "myapp:latest",
				"repository":   "my-repo",
				"auth": map[string]any{
					"method":   "service_account",
					"key_file": "/path/to/key.json",
				},
			},
			wantErrors: 0,
		},
		{
			name: "valid legacy GCR config",
			config: map[string]any{
				"project":          "my-project",
				"image":            "my-app",
				"source_image":     "myapp:latest",
				"artifact_registry": false,
				"region":           "us",
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &GCRPlugin{}
			resp, err := p.Validate(context.Background(), tt.config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(resp.Errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.wantErrors, len(resp.Errors), resp.Errors)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	p := &GCRPlugin{}

	config := map[string]any{
		"project":          "my-project",
		"region":           "us-west1",
		"repository":       "my-repo",
		"image":            "my-app",
		"source_image":     "myapp:v1.0.0",
		"artifact_registry": true,
		"tags":             []string{"v1.0.0", "latest"},
		"auth": map[string]any{
			"method":   "service_account",
			"key_file": "/path/to/key.json",
		},
		"multi_region": map[string]any{
			"enabled": true,
			"regions": []string{"us", "europe", "asia"},
		},
		"dry_run": true,
	}

	cfg := p.parseConfig(config)

	if cfg.Project != "my-project" {
		t.Errorf("expected project 'my-project', got '%s'", cfg.Project)
	}

	if cfg.Region != "us-west1" {
		t.Errorf("expected region 'us-west1', got '%s'", cfg.Region)
	}

	if cfg.Repository != "my-repo" {
		t.Errorf("expected repository 'my-repo', got '%s'", cfg.Repository)
	}

	if cfg.Image != "my-app" {
		t.Errorf("expected image 'my-app', got '%s'", cfg.Image)
	}

	if cfg.SourceImage != "myapp:v1.0.0" {
		t.Errorf("expected source_image 'myapp:v1.0.0', got '%s'", cfg.SourceImage)
	}

	if !cfg.ArtifactRegistry {
		t.Error("expected artifact_registry to be true")
	}

	if len(cfg.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(cfg.Tags))
	}

	if !cfg.DryRun {
		t.Error("expected dry_run to be true")
	}
}

func TestParseConfigDefaults(t *testing.T) {
	p := &GCRPlugin{}

	cfg := p.parseConfig(map[string]any{})

	if !cfg.ArtifactRegistry {
		t.Error("expected artifact_registry to default to true")
	}

	if cfg.Region != "us-central1" {
		t.Errorf("expected region to default to 'us-central1', got '%s'", cfg.Region)
	}

	if cfg.AuthMethod != "gcloud" {
		t.Errorf("expected auth method to default to 'gcloud', got '%s'", cfg.AuthMethod)
	}

	if len(cfg.Tags) != 1 || cfg.Tags[0] != "{{.Version}}" {
		t.Errorf("expected default tag ['{{.Version}}'], got %v", cfg.Tags)
	}
}

func TestProcessTags(t *testing.T) {
	p := &GCRPlugin{}

	ctx := &plugin.ReleaseContext{
		Version:         "1.2.3",
		PreviousVersion: "1.2.2",
		TagName:         "v1.2.3",
		Branch:          "feature/new-feature",
		ReleaseType:     "minor",
	}

	tests := []struct {
		name     string
		tags     []string
		expected []string
	}{
		{
			name:     "version tag",
			tags:     []string{"{{.Version}}"},
			expected: []string{"1.2.3"},
		},
		{
			name:     "multiple tags",
			tags:     []string{"{{.Version}}", "latest", "{{.TagName}}"},
			expected: []string{"1.2.3", "latest", "v1.2.3"},
		},
		{
			name:     "branch tag",
			tags:     []string{"{{.Branch}}"},
			expected: []string{"feature-new-feature"},
		},
		{
			name:     "conditional tag skipped",
			tags:     []string{"{{if not .Prerelease}}latest{{end}}"},
			expected: []string{}, // Conditional templates are skipped
		},
		{
			name:     "literal tag",
			tags:     []string{"stable"},
			expected: []string{"stable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.processTags(tt.tags, ctx)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("tag[%d]: expected '%s', got '%s'", i, tt.expected[i], tag)
				}
			}
		})
	}
}

func TestProcessTemplate(t *testing.T) {
	p := &GCRPlugin{}

	tests := []struct {
		name     string
		template string
		ctx      *plugin.ReleaseContext
		expected string
	}{
		{
			name:     "version only",
			template: "{{.Version}}",
			ctx:      &plugin.ReleaseContext{Version: "2.0.0"},
			expected: "2.0.0",
		},
		{
			name:     "mixed template",
			template: "v{{.Version}}-{{.ReleaseType}}",
			ctx:      &plugin.ReleaseContext{Version: "1.0.0", ReleaseType: "patch"},
			expected: "v1.0.0-patch",
		},
		{
			name:     "no placeholders",
			template: "latest",
			ctx:      &plugin.ReleaseContext{Version: "1.0.0"},
			expected: "latest",
		},
		{
			name:     "branch with slashes",
			template: "{{.Branch}}",
			ctx:      &plugin.ReleaseContext{Branch: "feature/foo/bar"},
			expected: "feature-foo-bar",
		},
		{
			name:     "conditional skipped",
			template: "{{if .Prerelease}}pre{{end}}",
			ctx:      &plugin.ReleaseContext{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.processTemplate(tt.template, tt.ctx)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
