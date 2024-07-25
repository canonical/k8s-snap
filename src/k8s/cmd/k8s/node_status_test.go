package k8s_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s"
	cmdutil "github.com/canonical/k8s/cmd/util"
	k8sdmock "github.com/canonical/k8s/pkg/client/k8sd/mock"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/lxd/shared/api"

	. "github.com/onsi/gomega"
)

func TestGetNodeStatusUtil(t *testing.T) {
	type testcase struct {
		expectedCode     int
		name             string
		nodeStatusErr    error
		expectedInStdErr []string
		nodeStatusResult apiv1.NodeStatus
	}

	tests := []testcase{
		{
			name: "NoError",
			nodeStatusResult: apiv1.NodeStatus{
				Name: "name", Address: "addr",
				ClusterRole: apiv1.ClusterRoleControlPlane, DatastoreRole: apiv1.DatastoreRoleVoter,
			},
		},
		{
			name:             "DaemonNotInitialized",
			nodeStatusErr:    api.StatusErrorf(http.StatusServiceUnavailable, "Daemon not yet initialized"),
			expectedCode:     1,
			expectedInStdErr: []string{"The node is not part of a Kubernetes cluster. You can bootstrap a new cluster"},
		},
		{
			name:             "RandomError",
			nodeStatusErr:    errors.New("something went bad"),
			expectedCode:     1,
			expectedInStdErr: []string{"something went bad", "Failed to retrieve the node status"},
		},
		{
			name:             "ContextCanceled",
			nodeStatusErr:    context.Canceled,
			expectedCode:     1,
			expectedInStdErr: []string{context.Canceled.Error(), "Failed to retrieve the node status"},
		},
		{
			name:             "ContextDeadlineExceeded",
			nodeStatusErr:    context.DeadlineExceeded,
			expectedCode:     1,
			expectedInStdErr: []string{context.DeadlineExceeded.Error(), "Failed to retrieve the node status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			var (
				returnCode int
				stdout     = &bytes.Buffer{}
				stderr     = &bytes.Buffer{}
				mockClient = &k8sdmock.Mock{
					NodeStatusErr:    tt.nodeStatusErr,
					NodeStatusResult: tt.nodeStatusResult,
				}
				env = cmdutil.ExecutionEnvironment{
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
				cmd = k8s.NewRootCmd(env)
			)

			status := k8s.GetNodeStatus(mockClient, cmd, env)

			g.Expect(returnCode).To(Equal(tt.expectedCode))
			for _, exp := range tt.expectedInStdErr {
				g.Expect(stderr.String()).To(ContainSubstring(exp))
			}

			if tt.expectedCode == 0 {
				g.Expect(status).To(Equal(tt.nodeStatusResult))
			}
		})
	}
}
