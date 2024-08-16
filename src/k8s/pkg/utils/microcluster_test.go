package utils_test

import (
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestMicroclusterTimeout(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		g := NewWithT(t)

		m := map[string]string{}
		g.Expect(utils.MicroclusterTimeoutFromMap(m)).To(BeZero())
	})

	t.Run("Normal", func(t *testing.T) {
		g := NewWithT(t)

		timeout := 5 * time.Second
		m := map[string]string{}

		mWithTimeout := utils.MicroclusterMapWithTimeout(m, timeout)
		g.Expect(utils.MicroclusterTimeoutFromMap(mWithTimeout)).To(Equal(timeout))
	})
}
