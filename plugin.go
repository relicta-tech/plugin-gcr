package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/relicta-tech/relicta-plugin-sdk/helpers"
	"github.com/relicta-tech/relicta-plugin-sdk/plugin"
)

// Version is set at build time.
var Version = "dev"

// GCRPlugin implements the Relicta plugin interface for Google Container Registry.
type GCRPlugin struct{}

// Config holds the plugin configuration.
type Config struct {
	// GCP Configuration
	ArtifactRegistry bool
	Project          string
	Region           string
	Repository       string
	Image            string

	// Authentication
	AuthMethod string
	KeyFile    string
	KeyJSON    string

	// Source image
	SourceImage string

	// Tags
	Tags []string

	// Multi-region
	MultiRegionEnabled bool
	MultiRegionRegions []string

	// Behavior
	DryRun bool
}

// GetInfo returns plugin metadata.
func (p *GCRPlugin) GetInfo() plugin.Info {
	return plugin.Info{
		Name:        "gcr",
		Version:     Version,
		Description: "Push container images to Google Artifact Registry and Google Container Registry",
		Hooks: []plugin.Hook{
			plugin.HookPostPublish,
		},
	}
}

// Validate validates the plugin configuration.
func (p *GCRPlugin) Validate(ctx context.Context, config map[string]any) (*plugin.ValidateResponse, error) {
	vb := helpers.NewValidationBuilder()
	cfg := p.parseConfig(config)

	// Project is required
	if cfg.Project == "" {
		vb.AddError("project", "GCP project ID is required")
	}

	// Image name is required
	if cfg.Image == "" {
		vb.AddError("image", "image name is required")
	}

	// Source image is required
	if cfg.SourceImage == "" {
		vb.AddError("source_image", "source image is required")
	}

	// Repository required for Artifact Registry
	if cfg.ArtifactRegistry && cfg.Repository == "" {
		vb.AddError("repository", "repository name required for Artifact Registry")
	}

	// Validate auth method
	if cfg.AuthMethod != "" && cfg.AuthMethod != "gcloud" && cfg.AuthMethod != "service_account" {
		vb.AddError("auth.method", "auth method must be 'gcloud' or 'service_account'")
	}

	// Service account requires key
	if cfg.AuthMethod == "service_account" && cfg.KeyFile == "" && cfg.KeyJSON == "" {
		vb.AddError("auth", "service account requires key_file or key_json")
	}

	return vb.Build(), nil
}

// Execute runs the plugin logic.
func (p *GCRPlugin) Execute(ctx context.Context, req plugin.ExecuteRequest) (*plugin.ExecuteResponse, error) {
	cfg := p.parseConfig(req.Config)
	cfg.DryRun = cfg.DryRun || req.DryRun

	// Process tag templates
	tags := p.processTags(cfg.Tags, &req.Context)

	// Create GCR client
	client := NewGCRClient(&GCRConfig{
		Project:          cfg.Project,
		Region:           cfg.Region,
		Repository:       cfg.Repository,
		ArtifactRegistry: cfg.ArtifactRegistry,
	})

	// Determine regions to push to
	regions := []string{cfg.Region}
	if cfg.MultiRegionEnabled && len(cfg.MultiRegionRegions) > 0 {
		regions = cfg.MultiRegionRegions
	}

	// Authenticate with GCR
	if !cfg.DryRun {
		for _, region := range regions {
			authCfg := &AuthConfig{
				Method:  cfg.AuthMethod,
				KeyFile: cfg.KeyFile,
				KeyJSON: cfg.KeyJSON,
			}
			if err := client.Authenticate(ctx, region, authCfg); err != nil {
				return nil, fmt.Errorf("failed to authenticate with %s: %w", region, err)
			}
		}
	}

	// Create Docker client
	docker := NewDockerClient()

	// Push images to each region
	pushedImages := []string{}
	for _, region := range regions {
		regionClient := NewGCRClient(&GCRConfig{
			Project:          cfg.Project,
			Region:           region,
			Repository:       cfg.Repository,
			ArtifactRegistry: cfg.ArtifactRegistry,
		})

		for _, tag := range tags {
			if tag == "" {
				continue
			}

			targetImage := fmt.Sprintf("%s:%s", regionClient.GetImagePath(cfg.Image), tag)

			if cfg.DryRun {
				fmt.Printf("[dry-run] Would tag %s as %s\n", cfg.SourceImage, targetImage)
				fmt.Printf("[dry-run] Would push %s\n", targetImage)
			} else {
				// Tag the image
				if err := docker.Tag(ctx, cfg.SourceImage, targetImage); err != nil {
					return nil, fmt.Errorf("failed to tag image: %w", err)
				}

				// Push the image
				if err := docker.Push(ctx, targetImage); err != nil {
					return nil, fmt.Errorf("failed to push image: %w", err)
				}

				fmt.Printf("Pushed: %s\n", targetImage)
			}

			pushedImages = append(pushedImages, targetImage)
		}
	}

	return &plugin.ExecuteResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully pushed %d image(s) to GCR", len(pushedImages)),
		Outputs: map[string]any{
			"project":       cfg.Project,
			"repository":    cfg.Repository,
			"tags":          tags,
			"pushed_images": pushedImages,
		},
	}, nil
}

// parseConfig parses the raw configuration into a Config struct.
func (p *GCRPlugin) parseConfig(raw map[string]any) *Config {
	parser := helpers.NewConfigParser(raw)

	tags := parser.GetStringSlice("tags", nil)
	if len(tags) == 0 {
		tags = []string{"{{.Version}}"}
	}

	multiRegionRegions := parser.GetStringSlice("multi_region.regions", nil)

	// Parse nested auth config
	authMethod := "gcloud"
	keyFile := ""
	keyJSON := ""
	if authRaw, ok := raw["auth"].(map[string]any); ok {
		authParser := helpers.NewConfigParser(authRaw)
		authMethod = authParser.GetString("method", "", "gcloud")
		keyFile = authParser.GetString("key_file", "GOOGLE_APPLICATION_CREDENTIALS", "")
		keyJSON = authParser.GetString("key_json", "GCP_SERVICE_ACCOUNT_JSON", "")
	}

	// Parse nested multi_region config
	multiRegionEnabled := false
	if mrRaw, ok := raw["multi_region"].(map[string]any); ok {
		mrParser := helpers.NewConfigParser(mrRaw)
		multiRegionEnabled = mrParser.GetBool("enabled", false)
	}

	return &Config{
		// GCP Configuration
		ArtifactRegistry: parser.GetBool("artifact_registry", true),
		Project:          parser.GetString("project", "CLOUDSDK_CORE_PROJECT", ""),
		Region:           parser.GetString("region", "", "us-central1"),
		Repository:       parser.GetString("repository", "", ""),
		Image:            parser.GetString("image", "", ""),

		// Authentication
		AuthMethod: authMethod,
		KeyFile:    keyFile,
		KeyJSON:    keyJSON,

		// Source image
		SourceImage: parser.GetString("source_image", "", ""),

		// Tags
		Tags: tags,

		// Multi-region
		MultiRegionEnabled: multiRegionEnabled,
		MultiRegionRegions: multiRegionRegions,

		// Behavior
		DryRun: parser.GetBool("dry_run", false),
	}
}

// processTags processes tag templates with release context.
func (p *GCRPlugin) processTags(tags []string, ctx *plugin.ReleaseContext) []string {
	processed := make([]string, 0, len(tags))

	for _, tag := range tags {
		result := p.processTemplate(tag, ctx)
		if result != "" {
			processed = append(processed, result)
		}
	}

	return processed
}

// processTemplate replaces template variables with actual values.
func (p *GCRPlugin) processTemplate(tmpl string, ctx *plugin.ReleaseContext) string {
	result := tmpl

	// Handle conditional templates (simplified)
	if strings.Contains(result, "{{if") {
		// Skip conditional templates for now - return empty
		return ""
	}

	// Replace common template variables
	result = strings.ReplaceAll(result, "{{.Version}}", ctx.Version)
	result = strings.ReplaceAll(result, "{{.PreviousVersion}}", ctx.PreviousVersion)
	result = strings.ReplaceAll(result, "{{.TagName}}", ctx.TagName)
	result = strings.ReplaceAll(result, "{{.ReleaseType}}", ctx.ReleaseType)

	// Handle branch name
	if ctx.Branch != "" {
		safeBranch := strings.ReplaceAll(ctx.Branch, "/", "-")
		result = strings.ReplaceAll(result, "{{.Branch}}", safeBranch)
	}

	return result
}
