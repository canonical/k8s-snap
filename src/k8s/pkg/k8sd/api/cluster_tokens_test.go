package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/snap"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/microcluster/v2/microcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// MockProvider is a manual mock implementation of the Provider interface.
type MockProvider struct {
	mockSnap snap.Snap
}

func (m *MockProvider) MicroCluster() *microcluster.MicroCluster {
	return nil
}

func (m *MockProvider) Snap() snap.Snap {
	return m.mockSnap
}

func (m *MockProvider) NotifyUpdateNodeConfigController() {}

func (m *MockProvider) NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) {
}

func TestPostClusterJoinTokens_DuplicateNode(t *testing.T) {
	// Setup
	mockSnap := &snapmock.Snap{}

	// Create fake K8s client with existing node
	existingNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "duplicate-node",
		},
	}
	fakeClientset := fake.NewSimpleClientset(existingNode)
	k8sClient := &kubernetes.Client{
		Interface: fakeClientset,
	}

	// Configure mockSnap to return our fake k8s client
	// The manual mock implementation in pkg/snap/mock uses the Mock struct field
	mockSnap.Mock.KubernetesClient = k8sClient

	mockProvider := &MockProvider{
		mockSnap: mockSnap,
	}

	endpoints := &Endpoints{
		context:  context.Background(),
		provider: mockProvider,
	}

	// Request with duplicate name
	reqBody := apiv1.GetJoinTokenRequest{
		Name: "duplicate-node",
		TTL:  time.Hour,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/cluster/tokens", bytes.NewReader(bodyBytes))

	// Execute
	// Passing nil for state.State strictly because the duplicate node check
	// happens before state is used.
	resp := endpoints.postClusterJoinTokens(nil, req)

	// Verify response
	w := httptest.NewRecorder()
	err := resp.Render(w, req)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Check error message in body
	var respBody map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	expectedError := "A node with the same name \"duplicate-node\" is already part of the cluster"
	if respBody["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, respBody["error"])
	}
}
