package main

import (
	"testing"
)

func TestNewGCRClient(t *testing.T) {
	config := &GCRConfig{
		Project:          "my-project",
		Region:           "us-central1",
		Repository:       "my-repo",
		ArtifactRegistry: true,
	}

	client := NewGCRClient(config)
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.config.Project != "my-project" {
		t.Errorf("expected project 'my-project', got '%s'", client.config.Project)
	}
}

func TestGetImagePath(t *testing.T) {
	tests := []struct {
		name     string
		config   *GCRConfig
		image    string
		expected string
	}{
		{
			name: "artifact registry",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "us-central1",
				Repository:       "my-repo",
				ArtifactRegistry: true,
			},
			image:    "my-app",
			expected: "us-central1-docker.pkg.dev/my-project/my-repo/my-app",
		},
		{
			name: "artifact registry europe",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "europe-west1",
				Repository:       "my-repo",
				ArtifactRegistry: true,
			},
			image:    "my-app",
			expected: "europe-west1-docker.pkg.dev/my-project/my-repo/my-app",
		},
		{
			name: "legacy GCR us",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "us",
				ArtifactRegistry: false,
			},
			image:    "my-app",
			expected: "gcr.io/my-project/my-app",
		},
		{
			name: "legacy GCR eu",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "eu",
				ArtifactRegistry: false,
			},
			image:    "my-app",
			expected: "eu.gcr.io/my-project/my-app",
		},
		{
			name: "legacy GCR europe",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "europe",
				ArtifactRegistry: false,
			},
			image:    "my-app",
			expected: "eu.gcr.io/my-project/my-app",
		},
		{
			name: "legacy GCR asia",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "asia",
				ArtifactRegistry: false,
			},
			image:    "my-app",
			expected: "asia.gcr.io/my-project/my-app",
		},
		{
			name: "legacy GCR default region",
			config: &GCRConfig{
				Project:          "my-project",
				Region:           "unknown",
				ArtifactRegistry: false,
			},
			image:    "my-app",
			expected: "gcr.io/my-project/my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGCRClient(tt.config)
			result := client.GetImagePath(tt.image)

			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetRegistryHost(t *testing.T) {
	tests := []struct {
		name     string
		config   *GCRConfig
		expected string
	}{
		{
			name: "artifact registry",
			config: &GCRConfig{
				Region:           "us-central1",
				ArtifactRegistry: true,
			},
			expected: "us-central1-docker.pkg.dev",
		},
		{
			name: "legacy GCR us",
			config: &GCRConfig{
				Region:           "us",
				ArtifactRegistry: false,
			},
			expected: "gcr.io",
		},
		{
			name: "legacy GCR eu",
			config: &GCRConfig{
				Region:           "eu",
				ArtifactRegistry: false,
			},
			expected: "eu.gcr.io",
		},
		{
			name: "legacy GCR asia",
			config: &GCRConfig{
				Region:           "asia",
				ArtifactRegistry: false,
			},
			expected: "asia.gcr.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGCRClient(tt.config)
			result := client.GetRegistryHost()

			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
