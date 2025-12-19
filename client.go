package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GCRConfig holds GCR client configuration.
type GCRConfig struct {
	Project          string
	Region           string
	Repository       string
	ArtifactRegistry bool
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Method  string
	KeyFile string
	KeyJSON string
}

// GCRClient provides GCR/Artifact Registry operations.
type GCRClient struct {
	config *GCRConfig
}

// NewGCRClient creates a new GCR client.
func NewGCRClient(config *GCRConfig) *GCRClient {
	return &GCRClient{
		config: config,
	}
}

// Authenticate authenticates with GCR/Artifact Registry.
func (c *GCRClient) Authenticate(ctx context.Context, region string, auth *AuthConfig) error {
	if auth == nil {
		auth = &AuthConfig{Method: "gcloud"}
	}

	switch auth.Method {
	case "gcloud", "":
		return c.authenticateGcloud(ctx, region)
	case "service_account":
		return c.authenticateServiceAccount(ctx, region, auth)
	default:
		return fmt.Errorf("unknown auth method: %s", auth.Method)
	}
}

// authenticateGcloud uses gcloud CLI for authentication.
func (c *GCRClient) authenticateGcloud(ctx context.Context, region string) error {
	registryHost := c.getRegistryHost(region)

	cmd := exec.CommandContext(ctx, "gcloud", "auth", "configure-docker", registryHost, "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gcloud auth failed: %w\n%s", err, string(output))
	}

	return nil
}

// authenticateServiceAccount uses service account for authentication.
func (c *GCRClient) authenticateServiceAccount(ctx context.Context, region string, auth *AuthConfig) error {
	registryHost := c.getRegistryHost(region)

	var keyData string
	if auth.KeyFile != "" {
		// Read key from file
		cmd := exec.CommandContext(ctx, "cat", auth.KeyFile)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
		keyData = string(output)
	} else if auth.KeyJSON != "" {
		keyData = auth.KeyJSON
	} else {
		return fmt.Errorf("service account key not provided")
	}

	// Docker login with service account
	cmd := exec.CommandContext(ctx, "docker", "login",
		"-u", "_json_key",
		"--password-stdin",
		registryHost,
	)
	cmd.Stdin = strings.NewReader(keyData)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker login failed: %w\n%s", err, string(output))
	}

	return nil
}

// getRegistryHost returns the registry host URL.
func (c *GCRClient) getRegistryHost(region string) string {
	if c.config.ArtifactRegistry {
		return fmt.Sprintf("%s-docker.pkg.dev", region)
	}

	// Legacy GCR
	switch region {
	case "us":
		return "gcr.io"
	case "eu", "europe":
		return "eu.gcr.io"
	case "asia":
		return "asia.gcr.io"
	default:
		return "gcr.io"
	}
}

// GetImagePath returns the full image path.
func (c *GCRClient) GetImagePath(image string) string {
	if c.config.ArtifactRegistry {
		return fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s",
			c.config.Region, c.config.Project, c.config.Repository, image)
	}

	registryHost := c.getRegistryHost(c.config.Region)
	return fmt.Sprintf("%s/%s/%s", registryHost, c.config.Project, image)
}

// GetRegistryHost returns the registry host for the current configuration.
func (c *GCRClient) GetRegistryHost() string {
	return c.getRegistryHost(c.config.Region)
}
