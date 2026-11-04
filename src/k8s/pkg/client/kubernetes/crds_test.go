package kubernetes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestCRDEstablished(t *testing.T) {
	tests := []struct {
		name           string
		crdName        string
		handler        http.HandlerFunc
		expectedResult bool
		expectedError  string
	}{
		{
			name:    "CRD is established",
			crdName: "upgrades.k8sd.io",
			handler: func(w http.ResponseWriter, r *http.Request) {
				crd := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: "upgrades.k8sd.io"},
					Status: apiextensionsv1.CustomResourceDefinitionStatus{
						Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
							{
								Type:   apiextensionsv1.Established,
								Status: apiextensionsv1.ConditionTrue,
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(crd)
			},
			expectedResult: true,
		},
		{
			name:    "CRD exists but not established",
			crdName: "upgrades.k8sd.io",
			handler: func(w http.ResponseWriter, r *http.Request) {
				crd := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: "upgrades.k8sd.io"},
					Status: apiextensionsv1.CustomResourceDefinitionStatus{
						Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
							{
								Type:   apiextensionsv1.Established,
								Status: apiextensionsv1.ConditionFalse,
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(crd)
			},
			expectedResult: false,
		},
		{
			name:    "CRD exists but no conditions",
			crdName: "upgrades.k8sd.io",
			handler: func(w http.ResponseWriter, r *http.Request) {
				crd := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: "upgrades.k8sd.io"},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(crd)
			},
			expectedResult: false,
		},
		{
			name:    "CRD does not exist",
			crdName: "upgrades.k8sd.io",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(metav1.Status{
					Status: metav1.StatusFailure,
					Reason: metav1.StatusReasonNotFound,
					Code:   http.StatusNotFound,
				})
			},
			expectedResult: false,
		},
		{
			name:    "API server error",
			crdName: "upgrades.k8sd.io",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: "failed to get CRD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := &Client{
				config: &rest.Config{Host: server.URL},
			}

			result, err := client.CRDEstablished(context.Background(), tt.crdName)

			if tt.expectedError != "" {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedError))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(result).To(Equal(tt.expectedResult))
			}
		})
	}
}
