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
	const tokenFileName = "tmp-token-file"
	const tokenHeaderName = "X-Node-Token"

	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

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
			name:               "missing header token",
			tokenHeaderContent: "",
			tokenFileContent:   "node-token",
			expectErr:          true,
			createFile:         true,
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
			tokenFileContent:   "",
			expectErr:          true,
			createFile:         true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			var err error
			if tc.createFile {
				err = os.WriteFile(path.Join(tempDir, tokenFileName), []byte(tc.tokenFileContent), 0o644)
				g.Expect(err).To(BeNil())
			}

			e := &Endpoints{
				context: context.Background(),
				provider: &mock.Provider{
					SnapFn: func() snap.Snap {
						return &mock.Snap{
							Mock: mock.Mock{
								NodeTokenFile: path.Join(tempDir, tokenFileName),
							},
						}
					},
				},
			}

			req := &http.Request{
				Header: make(http.Header),
			}
			req.Header.Set(tokenHeaderName, tc.tokenHeaderContent)

			handler := e.ValidateNodeTokenAccessHandler(tokenHeaderName)
			valid, resp := handler(nil, req)

			os.Remove(path.Join(tempDir, tokenFileName))

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
