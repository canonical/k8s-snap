package k8s_test

import (
	"testing"

	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"

	"github.com/canonical/k8s/pkg/utils/k8s"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListResourcesForGroupVersion(t *testing.T) {
	tests := []struct {
		name          string
		groupVersion  string
		expectedList  *v1.APIResourceList
		expectedError bool
	}{
		{
			name:         "Success scenario",
			groupVersion: "cilium.io/v2alpha1",
			expectedList: &v1.APIResourceList{
				GroupVersion: "cilium.io/v2alpha1",
				APIResources: []v1.APIResource{
					{Name: "test"},
				},
			},
			expectedError: false,
		},
		{
			name:          "Failure scenario",
			groupVersion:  "cilium.io/v2alpha1",
			expectedList:  nil,
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			clientset := fakeclientset.NewSimpleClientset()
			fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
			g.Expect(ok).To(BeTrue())

			if tc.expectedList != nil {
				fakeDiscovery.Resources = []*v1.APIResourceList{tc.expectedList}
			}

			// Create a new k8s client with the fake discovery client
			client := &k8s.Client{
				Interface: clientset,
			}

			// Call the ListResourcesForGroupVersion method
			resources, err := client.ListResourcesForGroupVersion(tc.groupVersion)

			if tc.expectedError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(resources).To(Equal(tc.expectedList))
			}
		})
	}
}
