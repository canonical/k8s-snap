package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/onsi/gomega"
)

func TestDeeplyUnwrapError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("when error is not wrapped", func(t *testing.T) {
		err := errors.New("test error")
		unwrapped := DeeplyUnwrapError(err)

		g.Expect(unwrapped).To(gomega.Equal(err))
	})

	t.Run("when error is wrapped once", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := fmt.Errorf("outer wrapper: %w", innerErr)

		unwrapped := DeeplyUnwrapError(err)

		g.Expect(unwrapped).To(gomega.Equal(innerErr))
	})

	t.Run("when error is deeply nested", func(t *testing.T) {
		innermostErr := errors.New("innermost error")
		innerErr := fmt.Errorf("middle wrapper: %w", innermostErr)
		err := fmt.Errorf("outer wrapper: %w", innerErr)

		unwrapped := DeeplyUnwrapError(err)

		g.Expect(unwrapped).To(gomega.Equal(innermostErr))
	})

	t.Run("when error is nil", func(t *testing.T) {
		var err error
		unwrapped := DeeplyUnwrapError(err)

		g.Expect(unwrapped).To(gomega.BeNil())
	})
}
