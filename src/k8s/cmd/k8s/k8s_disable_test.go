package k8s_test

import (
	"bytes"
	"context"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/k8s/client/mock"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestDisableCmd(t *testing.T) {
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
					Gateway: &apiv1.GatewayConfig{Enabled: vals.Pointer(false)},
				},
			},
		},
		{
			name:  "multiple",
			funcs: []string{"load-balancer", "gateway"},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway:      &apiv1.GatewayConfig{Enabled: vals.Pointer(false)},
					LoadBalancer: &apiv1.LoadBalancerConfig{Enabled: vals.Pointer(false)},
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

			stdout := &bytes.Buffer{}
			mockClient := &mock.Client{}
			env := cmdutil.ExecutionEnvironment{
				Stdout: stdout,
				Getuid: func() int { return 0 },
				Client: func(ctx context.Context) (client.Client, error) {
					return mockClient, nil
				},
			}
			cmd := k8s.NewRootCmd(env)

			cmd.SetArgs(append([]string{"disable"}, tt.funcs...))
			err := cmd.Execute()

			if tt.expectedErrMsg == "" {
				g.Expect(err).To(BeNil())
				g.Expect(stdout.String()).To(ContainSubstring("disabled"))
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrMsg))
			}
			g.Expect(mockClient.UpdateClusterConfigCalledWith).To(Equal(tt.expectedCall))
		})
	}
}
