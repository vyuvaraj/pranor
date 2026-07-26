package distribution

import (
	"strings"
	"testing"
)

func TestDistributionGenerator_DockerComposeAndHelm(t *testing.T) {
	gen := NewDistributionGenerator()

	components := []ComponentSpec{
		{Name: "servgate", Image: "servverse/servgate:latest", Port: 8080, Replicas: 2, EnvVars: map[string]string{"LOG_LEVEL": "info"}},
		{Name: "servqueue", Image: "servverse/servqueue:latest", Port: 9090, Replicas: 3},
		{Name: "servstore", Image: "servverse/servstore:latest", Port: 7070, Replicas: 1},
	}

	// Docker Compose
	compose := gen.GenerateDockerCompose(components)
	if compose.Format != FormatDockerCompose {
		t.Errorf("expected docker-compose format, got %s", compose.Format)
	}
	if !strings.Contains(compose.Content, "servgate:") || !strings.Contains(compose.Content, "servverse/servgate:latest") {
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
	if !strings.Contains(helm.Content, "replicas: 3") || !strings.Contains(helm.Content, "servqueue:") {
		t.Errorf("unexpected helm values content: %s", helm.Content)
	}
}
