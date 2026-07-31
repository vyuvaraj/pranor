package import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBlueGreenManager_DeployAndCutover(t *testing.T) {
	// Active Blue server
	blueSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer blueSrv.Close()

	// Healthy Green server
	greenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer greenSrv.Close()

	// Pranor Gate Mock
	gateSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer gateSrv.Close()

	cfg := BlueGreenConfig{
		ServiceName:     "user-auth-service",
		GatewayURL:      gateSrv.URL,
		BlueURL:         blueSrv.URL,
		GreenURL:        greenSrv.URL,
		ActiveSlot:      TargetBlue,
		HealthCheckPath: "/",
	}

	bgm := NewBlueGreenManager(cfg)

	if bgm.GetActiveSlot() != TargetBlue {
		t.Errorf("expected initial active slot Blue, got %s", bgm.GetActiveSlot())
	}

	// Trigger deployment to Green
	ctx := context.Background()
	newSlot, err := bgm.DeployNewVersion(ctx)
	if err != nil {
		t.Fatalf("DeployNewVersion failed: %v", err)
	}

	if newSlot != TargetGreen || bgm.GetActiveSlot() != TargetGreen {
		t.Errorf("expected active slot Green after deployment, got %s", newSlot)
	}
}

func TestBlueGreenManager_HealthCheckFailureAbortsCutover(t *testing.T) {
	blueSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer blueSrv.Close()

	// Failing Green server
	failingGreenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failingGreenSrv.Close()

	cfg := BlueGreenConfig{
		ServiceName:     "payment-svc",
		BlueURL:         blueSrv.URL,
		GreenURL:        failingGreenSrv.URL,
		ActiveSlot:      TargetBlue,
		HealthCheckPath: "/",
	}

	bgm := NewBlueGreenManager(cfg)

	_, err := bgm.DeployNewVersion(context.Background())
	if err == nil {
		t.Error("expected error when deploying to unhealthy target slot")
	}

	if bgm.GetActiveSlot() != TargetBlue {
		t.Errorf("expected active slot to remain Blue after aborted deployment, got %s", bgm.GetActiveSlot())
	}
}
