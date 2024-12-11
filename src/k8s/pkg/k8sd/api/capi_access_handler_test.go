package api_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/k8sd/database"
	microcluster_testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

func TestValidateCAPIAuthTokenAccessHandler(t *testing.T) {
	g := NewWithT(t)

	for _, tc := range []struct {
		name               string
		tokenHeaderContent string
		tokenDBContent     string
		expectErr          bool
	}{
		{
			name:               "valid token",
			tokenHeaderContent: "test-token",
			tokenDBContent:     "test-token",
			expectErr:          false,
		},
		{
			name:               "wrong token in header",
			tokenHeaderContent: "invalid-token",
			tokenDBContent:     "expected-token",
			expectErr:          true,
		},
		{
			name:               "wrong token in db",
			tokenHeaderContent: "expected-token",
			tokenDBContent:     "invalid-token",
			expectErr:          true,
		},
		{
			name:               "empty token in header",
			tokenHeaderContent: "",
			tokenDBContent:     "test-token",
			expectErr:          true,
		},
		{
			name:               "empty token in db",
			tokenHeaderContent: "test-token",
			tokenDBContent:     "",
			expectErr:          true,
		},
		{
			name:               "empty token in header and db",
			tokenHeaderContent: "",
			tokenDBContent:     "",
			expectErr:          true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			microcluster_testenv.WithState(t, func(ctx context.Context, s state.State) {
				var err error
				if tc.tokenDBContent != "" {
					err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
						return database.SetClusterAPIToken(ctx, tx, tc.tokenDBContent)
					})
					g.Expect(err).To(Not(HaveOccurred()))
				}

				req := &http.Request{
					Header: make(http.Header),
				}
				req.Header.Set(api.CAPITokenHeaderName, tc.tokenHeaderContent)

				handler := api.ValidateCAPIAuthTokenAccessHandler(api.CAPITokenHeaderName)
				valid, resp := handler(s, req)

				if tc.expectErr {
					g.Expect(valid).To(BeFalse())
					g.Expect(resp).To(Not(BeNil()))
				} else {
					g.Expect(valid).To(BeTrue())
					g.Expect(resp).To(BeNil())
				}
			})
		})
	}
}
