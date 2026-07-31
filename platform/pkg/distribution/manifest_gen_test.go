package import (
	"strings"
	"testing"
)

func TestDistributionGenerator_DockerComposeAndHelm(t *testing.T) {
	gen := NewDistributionGenerator()

	components := []ComponentSpec{
		{Name: "Pranor Gate", Image: "Pranor/Pranor Gate:latest", Port: 8080, Replicas: 2, EnvVars: map[string]string{"LOG_LEVEL": "info"}},
		{Name: "Pranor Pulse", Image: "Pranor/Pranor Pulse:latest", Port: 9090, Replicas: 3},
		{Name: "Pranor Vault", Image: "Pranor/Pranor Vault:latest", Port: 7070, Replicas: 1},
	}

	// Docker Compose
	compose := gen.GenerateDockerCompose(components)
	if compose.Format != FormatDockerCompose {
		t.Errorf("expected docker-compose format, got %s", compose.Format)
	}
	if !strings.Contains(compose.Content, "Pranor Gate:") || !strings.Contains(compose.Content, "Pranor/Pranor Gate:latest") {
		t.Errorf("unexpected docker-compose content: %s", compose.Content)
	}
	if !strings.Contains(compose.Content, "LOG_LEVEL: info") {
		t.Errorf("expected env vars in docker-compose content")
	}

	// Helm values
	helm := gen.GenerateHelmValues(components)
	if helm.Format != FormatHelm {
		t.Errorf("expected helm format, got %s", helm.Format)
	}
	if !strings.Contains(helm.Content, "replicas: 3") || !strings.Contains(helm.Content, "Pranor Pulse:") {
		t.Errorf("unexpected helm values content: %s", helm.Content)
	}
}
