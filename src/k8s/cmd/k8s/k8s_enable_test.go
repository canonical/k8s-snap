package k8s

import (
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestK8sEnableCmd(t *testing.T) {
	tests := []struct {
		name           string
		funcs          []string
		expectedCall   apiv1.UpdateClusterConfigRequest
		expectedErrMsg string
	}{
		{
			name:           "empty",
			funcs:          []string{},
			expectedErrMsg: "missing argument",
		},
		{
			name:  "one",
			funcs: []string{"gateway"},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway: &apiv1.GatewayConfig{Enabled: vals.Pointer(true)},
				},
			},
		},
		{
			name:  "multiple",
			funcs: []string{"load-balancer", "gateway"},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway:      &apiv1.GatewayConfig{Enabled: vals.Pointer(true)},
					LoadBalancer: &apiv1.LoadBalancerConfig{Enabled: vals.Pointer(true)},
				},
			},
		},
		{
			name:           "unknown",
			funcs:          []string{"unknownFunc"},
			expectedErrMsg: "unknown functionality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			cmd, client, out := mustSetupCLIWithFakeClient()

			// Disable root permission check.
			cmd.PersistentPreRunE = nil

			cmd.SetArgs(append([]string{"enable"}, tt.funcs...))
			err := cmd.Execute()

			if tt.expectedErrMsg == "" {
				g.Expect(err).To(BeNil())
				g.Expect(out).To(ContainSubstring("enabled"))
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrMsg))
			}
			g.Expect(client.UpdateClusterConfigCalledWith).To(Equal(tt.expectedCall))
		})
	}
}
