package snapdconfig_test

import (
	"context"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	. "github.com/onsi/gomega"
)

func TestParseMeta(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		g := NewWithT(t)
		s := &mock.Snap{
			Mock: mock.Mock{
				SnapctlGet: map[string][]byte{
					"-d meta": []byte(`{"meta":{"orb":"snapd","api-version":"1.30"}}`),
				},
			},
		}

		meta, empty, err := snapdconfig.ParseMeta(context.Background(), s)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(empty).To(BeTrue())
		g.Expect(meta).To(Equal(snapdconfig.Meta{Orb: "snapd", APIVersion: "1.30"}))
	})

	t.Run("Defaults", func(t *testing.T) {
		g := NewWithT(t)
		s := &mock.Snap{
			Mock: mock.Mock{
				SnapctlGet: map[string][]byte{
					"-d meta": []byte(`{"meta":{}}`),
				},
			},
		}

		meta, empty, err := snapdconfig.ParseMeta(context.Background(), s)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(empty).To(BeTrue())
		g.Expect(meta).To(Equal(snapdconfig.Meta{Orb: "none", APIVersion: "1.30"}))
	})

	t.Run("LegacyAPIVersionKeyIsIgnoredAndDefaulted", func(t *testing.T) {
		g := NewWithT(t)
		s := &mock.Snap{
			Mock: mock.Mock{
				SnapctlGet: map[string][]byte{
					"-d meta": []byte(`{"meta":{"orb":"snapd","apiVersion":"1.30"}}`),
				},
			},
		}

		meta, empty, err := snapdconfig.ParseMeta(context.Background(), s)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(empty).To(BeTrue())
		g.Expect(meta).To(Equal(snapdconfig.Meta{Orb: "snapd", APIVersion: "1.30"}))
	})

	t.Run("InvalidAPIVersion", func(t *testing.T) {
		g := NewWithT(t)
		s := &mock.Snap{
			Mock: mock.Mock{
				SnapctlGet: map[string][]byte{
					"-d meta": []byte(`{"meta":{"orb":"snapd","api-version":"2.0"}}`),
				},
			},
		}

		_, empty, err := snapdconfig.ParseMeta(context.Background(), s)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("meta.api-version"))
		g.Expect(empty).To(BeFalse())
	})
}

func TestSetMeta(t *testing.T) {
	g := NewWithT(t)
	s := &mock.Snap{}

	err := snapdconfig.SetMeta(context.Background(), s, snapdconfig.Meta{Orb: "snapd", APIVersion: "1.30"})
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(s.SnapctlSetCalledWith).To(HaveLen(1))
	g.Expect(s.SnapctlSetCalledWith[0]).To(ConsistOf(`meta={"orb":"snapd","api-version":"1.30"}`))
}
