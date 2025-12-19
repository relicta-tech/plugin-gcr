package main

import (
	"testing"
)

func TestNewDockerClient(t *testing.T) {
	client := NewDockerClient()
	if client == nil {
		t.Error("expected client, got nil")
	}
}
