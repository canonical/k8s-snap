package errors

import (
	"errors"
	"fmt"
	"testing"

	v1 "github.com/canonical/k8s/api/v1"
	. "github.com/onsi/gomega"
)

func TestResolve(t *testing.T) {
	g := NewGomegaWithT(t)

	mockErrorMsg := "unknown error: "
	mockExtraErrorMsgs := map[error]string{
		v1.ErrUnknown:         mockErrorMsg,
		v1.ErrNotBootstrapped: "not bootstrapped",
	}

	testCases := []struct {
		name        string
		err         error
		extraErrMsg map[error]string
		expected    error
	}{
		{
			name:     "NilError",
			err:      nil,
			expected: nil,
		},
		{
			name:     "UnknownError",
			err:      errors.New("unknown error"),
			expected: fmt.Errorf("%s%s", genericErrorMsgs[v1.ErrUnknown], "unknown error"),
		},
		{
			name:        "UnknownErrorWithCustomMsg",
			err:         errors.New("unknown error"),
			extraErrMsg: mockExtraErrorMsgs,
			expected:    fmt.Errorf("%s%s", mockExtraErrorMsgs[v1.ErrUnknown], "unknown error"),
		},
		{
			name:        "KnownErrorWithCustomMsg",
			err:         errors.New(v1.ErrNotBootstrapped.Error()),
			extraErrMsg: mockExtraErrorMsgs,
			expected:    errors.New("not bootstrapped"),
		},
		{
			name:     "UnknownErrorNoMatch",
			err:      errors.New("some other error occurred"),
			expected: errors.New("some other error occurred"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Transform(&tc.err, tc.extraErrMsg)
			if tc.expected == nil {
				g.Expect(tc.err).To(BeNil())
			} else {
				g.Expect(tc.err).To(MatchError(tc.expected))
			}
		})
	}
}
