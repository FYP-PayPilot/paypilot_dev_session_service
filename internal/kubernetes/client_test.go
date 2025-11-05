package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name          string
		helmChartPath string
		wantErr       bool
	}{
		{
			name:          "with custom helm chart path",
			helmChartPath: "/custom/path",
			wantErr:       false,
		},
		{
			name:          "with default helm chart path",
			helmChartPath: "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(logger, tt.helmChartPath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.log)
				if tt.helmChartPath == "" {
					assert.Equal(t, "./helm/dev-session-template", client.helmChart)
				} else {
					assert.Equal(t, tt.helmChartPath, client.helmChart)
				}
			}
		})
	}
}

func TestClient_ServiceEndpoints(t *testing.T) {
	// Test ServiceEndpoints structure
	endpoints := &ServiceEndpoints{
		PreviewURL:  "http://10.0.0.1/preview",
		PreviewPath: "/preview",
		ChatURL:     "http://10.0.0.1/chat",
		ChatPath:    "/chat",
		VscodeURL:   "http://10.0.0.1/vscode",
		VscodePath:  "/vscode",
		ClusterIP:   "10.0.0.1",
	}

	assert.Equal(t, "http://10.0.0.1/preview", endpoints.PreviewURL)
	assert.Equal(t, "/preview", endpoints.PreviewPath)
	assert.Equal(t, "http://10.0.0.1/chat", endpoints.ChatURL)
	assert.Equal(t, "/chat", endpoints.ChatPath)
	assert.Equal(t, "http://10.0.0.1/vscode", endpoints.VscodeURL)
	assert.Equal(t, "/vscode", endpoints.VscodePath)
	assert.Equal(t, "10.0.0.1", endpoints.ClusterIP)
}

func TestIsValidProjectUUID(t *testing.T) {
	tests := []struct {
		name  string
		uuid  string
		valid bool
	}{
		{
			name:  "valid UUID",
			uuid:  "550e8400-e29b-41d4-a716-446655440000",
			valid: true,
		},
		{
			name:  "valid UUID uppercase",
			uuid:  "550E8400-E29B-41D4-A716-446655440000",
			valid: true,
		},
		{
			name:  "invalid UUID - wrong format",
			uuid:  "550e8400-e29b-41d4-a716",
			valid: false,
		},
		{
			name:  "invalid UUID - command injection attempt",
			uuid:  "550e8400-e29b-41d4-a716-446655440000; rm -rf /",
			valid: false,
		},
		{
			name:  "invalid UUID - special characters",
			uuid:  "550e8400-e29b-41d4-a716-44665544000@",
			valid: false,
		},
		{
			name:  "empty UUID",
			uuid:  "",
			valid: false,
		},
		{
			name:  "invalid UUID - no hyphens",
			uuid:  "550e8400e29b41d4a716446655440000",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidProjectUUID(tt.uuid)
			assert.Equal(t, tt.valid, result, "UUID validation mismatch for: %s", tt.uuid)
		})
	}
}
