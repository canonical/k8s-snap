package k8s_test

import (
	"bytes"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s"
	cmdutil "github.com/canonical/k8s/cmd/util"
	k8sdmock "github.com/canonical/k8s/pkg/client/k8sd/mock"
	"github.com/canonical/k8s/pkg/k8sd/features"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
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
			funcs: []string{features.Gateway},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway: apiv1.GatewayConfig{Enabled: utils.Pointer(false)},
				},
			},
			expectedStdout: "disabled",
		},
		{
			name:  "multiple",
			funcs: []string{features.LoadBalancer, features.Gateway},
			expectedCall: apiv1.UpdateClusterConfigRequest{
				Config: apiv1.UserFacingClusterConfig{
					Gateway:      apiv1.GatewayConfig{Enabled: utils.Pointer(false)},
					LoadBalancer: apiv1.LoadBalancerConfig{Enabled: utils.Pointer(false)},
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
			mockClient := &k8sdmock.Mock{}
			var returnCode int
			env := cmdutil.ExecutionEnvironment{
				Stdout: stdout,
				Stderr: stderr,
				Getuid: func() int { return 0 },
				Snap: &snapmock.Snap{
					Mock: snapmock.Mock{
						K8sdClient: mockClient,
					},
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
				g.Expect(mockClient.SetClusterConfigCalledWith).To(Equal(tt.expectedCall))
			}
		})
	}
}
