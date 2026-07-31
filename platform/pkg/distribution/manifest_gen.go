package import (
	"encoding/json"
	"fmt"
	"strings"
)

// DistributionFormat specifies the target deployment distribution format.
type DistributionFormat string

const (
	FormatDockerCompose DistributionFormat = "docker-compose"
	FormatHelm          DistributionFormat = "helm"
)

// ComponentSpec defines a Pranor component deployment specification.
type ComponentSpec struct {
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Port       int               `json:"port"`
	Replicas   int               `json:"replicas"`
	EnvVars    map[string]string `json:"env_vars,omitempty"`
}

// DistributionManifest holds a rendered deployment manifest.
type DistributionManifest struct {
	Format  DistributionFormat `json:"format"`
	Content string             `json:"content"`
}

// DistributionGenerator renders Docker Compose and Helm chart manifests for Pranor platform deployment.
type DistributionGenerator struct{}

// NewDistributionGenerator creates a DistributionGenerator instance.
func NewDistributionGenerator() *DistributionGenerator {
	return &DistributionGenerator{}
}

// GenerateDockerCompose renders a docker-compose.yml for given components.
func (dg *DistributionGenerator) GenerateDockerCompose(components []ComponentSpec) DistributionManifest {
	var sb strings.Builder
	sb.WriteString("version: '3.8'\nservices:\n")

	for _, c := range components {
		sb.WriteString(fmt.Sprintf("  %s:\n", c.Name))
		sb.WriteString(fmt.Sprintf("    image: %s\n", c.Image))
		sb.WriteString(fmt.Sprintf("    ports:\n      - \"%d:%d\"\n", c.Port, c.Port))
		if len(c.EnvVars) > 0 {
			sb.WriteString("    environment:\n")
			for k, v := range c.EnvVars {
				sb.WriteString(fmt.Sprintf("      %s: %s\n", k, v))
			}
		}
	}

	return DistributionManifest{Format: FormatDockerCompose, Content: sb.String()}
}

// GenerateHelmValues renders a values.yaml for given components.
func (dg *DistributionGenerator) GenerateHelmValues(components []ComponentSpec) DistributionManifest {
	var sb strings.Builder
	sb.WriteString("# Pranor Production Helm Values\nglobal:\n  imageTag: latest\nservices:\n")

	for _, c := range components {
		replicas := c.Replicas
		if replicas <= 0 {
			replicas = 1
		}
		sb.WriteString(fmt.Sprintf("  %s:\n    enabled: true\n    replicas: %d\n    image: %s\n    port: %d\n", c.Name, replicas, c.Image, c.Port))
	}

	return DistributionManifest{Format: FormatHelm, Content: sb.String()}
}

// MarshalManifest serializes a DistributionManifest to JSON.
func (dg *DistributionGenerator) MarshalManifest(manifest DistributionManifest) string {
	b, _ := json.MarshalIndent(manifest, "", "  ")
	return string(b)
}
