package cmdutil_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	k8sdmock "github.com/canonical/k8s/pkg/client/k8sd/mock"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/lxd/shared/api"

	. "github.com/onsi/gomega"
)

func TestGetNodeStatusUtil(t *testing.T) {
	type testcase struct {
		expIsBootstrapped bool
		name              string
		nodeStatusErr     error
		nodeStatusResult  apiv1.NodeStatus
	}

	tests := []testcase{
		{
			name:              "NoError",
			expIsBootstrapped: true,
			nodeStatusResult: apiv1.NodeStatus{
				Name: "name", Address: "addr",
				ClusterRole: apiv1.ClusterRoleControlPlane, DatastoreRole: apiv1.DatastoreRoleVoter,
			},
		},
		{
			name:              "DaemonNotInitialized",
			nodeStatusErr:     api.StatusErrorf(http.StatusServiceUnavailable, "Daemon not yet initialized"),
			expIsBootstrapped: false,
		},
		{
			name:              "RandomError",
			nodeStatusErr:     errors.New("something went bad"),
			expIsBootstrapped: true,
		},
		{
			name:              "ContextCanceled",
			nodeStatusErr:     context.Canceled,
			expIsBootstrapped: true,
		},
		{
			name:              "ContextDeadlineExceeded",
			nodeStatusErr:     context.DeadlineExceeded,
			expIsBootstrapped: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			var (
				mockClient = &k8sdmock.Mock{
					NodeStatusErr:    test.nodeStatusErr,
					NodeStatusResult: test.nodeStatusResult,
				}
				env = cmdutil.ExecutionEnvironment{
					Getuid: func() int { return 0 },
					Snap: &snapmock.Snap{
						Mock: snapmock.Mock{
							K8sdClient: mockClient,
						},
					},
				}
			)

			status, isBootstrapped, err := cmdutil.GetNodeStatus(context.TODO(), mockClient, env)

			g.Expect(isBootstrapped).To(Equal(test.expIsBootstrapped))
			if test.nodeStatusErr == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(test.nodeStatusErr))
			}

			if err == nil {
				g.Expect(status).To(Equal(test.nodeStatusResult))
			}
		})
	}
}
