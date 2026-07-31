package platform

import (
	"encoding/json"
	"net/http/httptest"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestNewWireGuardMesh(t *testing.T) {
	mesh, err := NewWireGuardMesh("node-1", 51820, "10.0.0.1/32")
	if err != nil {
		t.Fatalf("Failed to create WireGuardMesh: %v", err)
	}
	if mesh.nodeID != "node-1" {
		t.Errorf("Expected node-1, got %s", mesh.nodeID)
	}
	if mesh.listenPort != 51820 {
		t.Errorf("Expected 51820, got %d", mesh.listenPort)
	}
	if mesh.publicKey == (WireGuardKey{}) {
		t.Error("Expected non-zero public key")
	}
	pubStr := mesh.publicKey.String()
	if pubStr == "" {
		t.Error("Expected non-empty public key string")
	}
}

func TestWireGuardMeshAddAndListPeers(t *testing.T) {
	mesh, err := NewWireGuardMesh("node-1", 51820, "10.0.0.1/32")
	if err != nil {
		t.Fatalf("Failed to create WireGuardMesh: %v", err)
	}

	mesh.AddPeerManually("node-2", "192.168.1.2:51820", "dummykey==", "10.0.0.2/32")
	mesh.AddPeerManually("node-3", "192.168.1.3:51820", "anotherkey==", "10.0.0.3/32")

	peers := mesh.ListPeers()
	if len(peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(peers))
	}
}

func TestWireGuardMeshHTTPGetPeers(t *testing.T) {
	mesh, err := NewWireGuardMesh("node-test", 51820, "10.99.0.1/32")
	if err != nil {
		t.Fatalf("Failed to create WireGuardMesh: %v", err)
	}
	mesh.AddPeerManually("node-99", "10.0.0.99:51820", "testpubkey==", "10.0.0.99/32")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mesh/peers", nil)
	w := httptest.NewRecorder()
	mesh.HandleMeshPeers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Invalid JSON response: %v", err)
	}
	if _, ok := resp["peers"]; !ok {
		t.Error("Expected 'peers' key in response")
	}
}

func TestWireGuardMeshHTTPAddPeer(t *testing.T) {
	mesh, err := NewWireGuardMesh("node-src", 51820, "10.0.1.1/32")
	if err != nil {
		t.Fatalf("Failed to create WireGuardMesh: %v", err)
	}

	body := `{"node_id":"node-dest","endpoint":"10.0.1.2:51820","public_key":"validkey==","allowed_ip":"10.0.1.2/32"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mesh/peers", strings.NewReader(body))
	w := httptest.NewRecorder()
	mesh.HandleMeshPeers(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", w.Code)
	}

	peers := mesh.ListPeers()
	if len(peers) != 1 || peers[0].NodeID != "node-dest" {
		t.Errorf("Expected peer node-dest, got: %+v", peers)
	}
}

func TestDerivePublicKey(t *testing.T) {
	var priv WireGuardKey
	priv[0] = 0x12
	priv[31] = 0x34

	pub1 := derivePublicKey(priv)
	pub2 := derivePublicKey(priv)

	if pub1 != pub2 {
		t.Error("derivePublicKey should be deterministic")
	}
	if pub1 == priv {
		t.Error("Public key should differ from private key")
	}
}

func TestDetectLocalIP(t *testing.T) {
	mesh, _ := NewWireGuardMesh("test", 51820, "10.0.0.1/32")
	ip := mesh.detectLocalIP()
	parsed := net.ParseIP(ip)
	if parsed == nil {
		t.Errorf("detectLocalIP returned invalid IP: %s", ip)
	}
}
