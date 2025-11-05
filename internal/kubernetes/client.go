package kubernetes

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// Client handles Kubernetes operations for dev containers
type Client struct {
	log       *zap.Logger
	helmChart string // Path to the Helm chart template
}

// NewClient creates a new Kubernetes client
func NewClient(log *zap.Logger, helmChartPath string) (*Client, error) {
	// Default Helm chart path if not provided
	if helmChartPath == "" {
		helmChartPath = "./helm/dev-session-template"
	}

	return &Client{
		log:       log,
		helmChart: helmChartPath,
	}, nil
}

// ServiceEndpoints holds the service endpoint information
type ServiceEndpoints struct {
	PreviewURL  string
	PreviewPath string
	ChatURL     string
	ChatPath    string
	VscodeURL   string
	VscodePath  string
	ClusterIP   string
}

// CreateDevContainer creates a new dev container in the specified namespace using Helm
func (c *Client) CreateDevContainer(ctx context.Context, projectUUID string, projectID int, userID int) (*ServiceEndpoints, error) {
	releaseName := fmt.Sprintf("dev-session-%s", projectUUID)

	c.log.Info("Creating dev container with Helm",
		zap.String("release", releaseName),
		zap.String("namespace", projectUUID),
		zap.Int("project_id", projectID),
		zap.Int("user_id", userID))

	// Build Helm install command
	args := []string{
		"install", releaseName,
		c.helmChart,
		"--set", fmt.Sprintf("project.uuid=%s", projectUUID),
		"--set", fmt.Sprintf("project.id=%d", projectID),
		"--set", fmt.Sprintf("user.id=%d", userID),
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.log.Error("Failed to install Helm chart",
			zap.Error(err),
			zap.String("output", string(output)))
		return nil, fmt.Errorf("helm install failed: %w, output: %s", err, string(output))
	}

	c.log.Info("Helm chart installed successfully",
		zap.String("release", releaseName),
		zap.String("output", string(output)))

	// Get service endpoints
	endpoints, err := c.GetServiceEndpoints(ctx, projectUUID, releaseName)
	if err != nil {
		c.log.Warn("Failed to get service endpoints", zap.Error(err))
		// Return default endpoints even if we can't fetch them
		endpoints = &ServiceEndpoints{
			PreviewPath: "/preview",
			ChatPath:    "/chat",
			VscodePath:  "/vscode",
		}
	}

	return endpoints, nil
}

// GetServiceEndpoints retrieves the service endpoints for a dev session
func (c *Client) GetServiceEndpoints(ctx context.Context, namespace string, releaseName string) (*ServiceEndpoints, error) {
	// Get ClusterIP of the load balancer service
	serviceName := fmt.Sprintf("%s-dev-session-template-lb", releaseName)

	cmd := exec.CommandContext(ctx, "kubectl", "get", "service", serviceName,
		"-n", namespace,
		"-o", "jsonpath={.spec.clusterIP}")

	output, err := cmd.CombinedOutput()
	clusterIP := strings.TrimSpace(string(output))

	if err != nil || clusterIP == "" {
		c.log.Warn("Failed to get ClusterIP", zap.Error(err))
		clusterIP = "pending"
	}

	endpoints := &ServiceEndpoints{
		ClusterIP:   clusterIP,
		PreviewURL:  fmt.Sprintf("http://%s/preview", clusterIP),
		PreviewPath: "/preview",
		ChatURL:     fmt.Sprintf("http://%s/chat", clusterIP),
		ChatPath:    "/chat",
		VscodeURL:   fmt.Sprintf("http://%s/vscode", clusterIP),
		VscodePath:  "/vscode",
	}

	return endpoints, nil
}

// DeleteDevContainer deletes a dev container from Kubernetes using Helm
func (c *Client) DeleteDevContainer(ctx context.Context, projectUUID string) error {
	releaseName := fmt.Sprintf("dev-session-%s", projectUUID)

	c.log.Info("Deleting dev container with Helm",
		zap.String("release", releaseName),
		zap.String("namespace", projectUUID))

	// Uninstall Helm release
	cmd := exec.CommandContext(ctx, "helm", "uninstall", releaseName, "-n", projectUUID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.log.Error("Failed to uninstall Helm chart",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("helm uninstall failed: %w", err)
	}

	c.log.Info("Helm chart uninstalled successfully", zap.String("release", releaseName))

	// Optionally delete the namespace
	// Note: This is commented out as namespace deletion might be handled separately
	// cmd = exec.CommandContext(ctx, "kubectl", "delete", "namespace", projectUUID)
	// cmd.Run()

	return nil
}

// GetContainerStatus returns the status of a dev container
func (c *Client) GetContainerStatus(ctx context.Context, namespace string, releaseName string) (string, error) {
	// Query Helm release status
	cmd := exec.CommandContext(ctx, "helm", "status", releaseName, "-n", namespace, "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.log.Error("Failed to get Helm release status",
			zap.Error(err),
			zap.String("output", string(output)))
		return "error", err
	}

	// Check if output contains "deployed" status
	if strings.Contains(string(output), `"status":"deployed"`) {
		return "running", nil
	} else if strings.Contains(string(output), `"status":"pending-install"`) {
		return "pending", nil
	} else if strings.Contains(string(output), `"status":"failed"`) {
		return "error", nil
	}

	return "unknown", nil
}

// UpdateContainer updates the configuration of a running dev container
func (c *Client) UpdateContainer(ctx context.Context, projectUUID string, projectID int, userID int) error {
	releaseName := fmt.Sprintf("dev-session-%s", projectUUID)

	c.log.Info("Updating dev container with Helm",
		zap.String("release", releaseName),
		zap.String("namespace", projectUUID))

	// Build Helm upgrade command
	args := []string{
		"upgrade", releaseName,
		c.helmChart,
		"-n", projectUUID,
		"--set", fmt.Sprintf("project.uuid=%s", projectUUID),
		"--set", fmt.Sprintf("project.id=%d", projectID),
		"--set", fmt.Sprintf("user.id=%d", userID),
		"--wait",
		"--timeout", "5m",
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.log.Error("Failed to upgrade Helm chart",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("helm upgrade failed: %w", err)
	}

	c.log.Info("Helm chart upgraded successfully", zap.String("release", releaseName))
	return nil
}
