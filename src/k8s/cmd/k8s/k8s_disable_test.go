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
		expectedCode   int
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "empty",
			funcs:          []string{},
			expectedStderr: "Error: requires at least 1 arg",
			expectedCode:   1,
		},
		{
			name:  "one",
			funcs: []string{"gateway"},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway: apiv1.GatewayConfig{Enabled: vals.Pointer(false)},
				},
			},
			expectedStdout: "disabled",
		},
		{
			name:  "multiple",
			funcs: []string{"load-balancer", "gateway"},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway:      apiv1.GatewayConfig{Enabled: vals.Pointer(false)},
					LoadBalancer: apiv1.LoadBalancerConfig{Enabled: vals.Pointer(false)},
				},
			},
			expectedStdout: "disabled",
		},
		{
			name:           "unknown",
			funcs:          []string{"unknownFunc"},
			expectedStderr: "Error: Cannot disable",
			expectedCode:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			mockClient := &mock.Client{}
			var returnCode int
			env := cmdutil.ExecutionEnvironment{
				Stdout: stdout,
				Stderr: stderr,
				Getuid: func() int { return 0 },
				Client: func(ctx context.Context) (client.Client, error) {
					return mockClient, nil
				},
				Exit: func(rc int) { returnCode = rc },
			}
			cmd := k8s.NewRootCmd(env)

			cmd.SetArgs(append([]string{"disable"}, tt.funcs...))
			cmd.Execute()

			g.Expect(stdout.String()).To(ContainSubstring(tt.expectedStdout))
			g.Expect(stderr.String()).To(ContainSubstring(tt.expectedStderr))
			g.Expect(returnCode).To(Equal(tt.expectedCode))

			if tt.expectedCode == 0 {
				g.Expect(mockClient.UpdateClusterConfigCalledWith).To(Equal(tt.expectedCall))
			}
		})
	}
}
