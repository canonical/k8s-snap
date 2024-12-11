package api

import (
	"context"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestValidateNodeTokenAccessHandler(t *testing.T) {
	for _, tc := range []struct {
		name               string
		tokenHeaderContent string
		tokenFileContent   string
		expectErr          bool
		createFile         bool
	}{
		{
			name:               "header and file token match",
			tokenHeaderContent: "node-token",
			tokenFileContent:   "node-token",
			createFile:         true,
		},
		{
			name:               "header and file token differ",
			tokenHeaderContent: "node-token",
			tokenFileContent:   "different-node-token",
			expectErr:          true,
			createFile:         true,
		},
		{
			name:             "missing header token",
			tokenFileContent: "node-token",
			expectErr:        true,
			createFile:       true,
		},
		{
			name:               "missing token file",
			tokenHeaderContent: "node-token",
			tokenFileContent:   "node-token",
			expectErr:          true,
		},
		{
			name:               "missing token in token file",
			tokenHeaderContent: "node-token",
			expectErr:          true,
			createFile:         true,
		},
		{
			name:      "missing header and token file",
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			dir := t.TempDir()

			var err error
			if tc.createFile {
				err = os.WriteFile(path.Join(dir, "token-file"), []byte(tc.tokenFileContent), 0o644)
				g.Expect(err).ToNot(HaveOccurred())
			}

			e := &Endpoints{
				context: context.Background(),
				provider: &mock.Provider{
					SnapFn: func() snap.Snap {
						return &mock.Snap{
							Mock: mock.Mock{
								NodeTokenFile: path.Join(dir, "token-file"),
							},
						}
					},
				},
			}

			req := &http.Request{
				Header: make(http.Header),
			}
			req.Header.Set(TokenHeaderName, tc.tokenHeaderContent)

			handler := e.ValidateNodeTokenAccessHandler(TokenHeaderName)
			valid, resp := handler(nil, req)

			if tc.expectErr {
				g.Expect(valid).To(BeFalse())
				g.Expect(resp).NotTo(BeNil())
			} else {
				g.Expect(valid).To(BeTrue())
				g.Expect(resp).To(BeNil())
			}
		})
	}
}
