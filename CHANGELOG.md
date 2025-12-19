# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-12-19

### Added

- Initial release
- Push container images to Google Artifact Registry
- Push container images to legacy Google Container Registry
- Authentication via gcloud CLI
- Authentication via service account key
- Multiple image tag support with template variables
- Tag templates: `{{.Version}}`, `{{.Branch}}`, `{{.TagName}}`, etc.
- Multi-region deployment support
- Dry-run mode for testing
- PostPublish hook support
