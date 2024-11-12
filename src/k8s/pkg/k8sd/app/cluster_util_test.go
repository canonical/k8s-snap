package app_test

import (
	"net/netip"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/app"
	mctypes "github.com/canonical/microcluster/v3/rest/types"
	. "github.com/onsi/gomega"
)

func TestDetermineLocalhostAddress(t *testing.T) {
	t.Run("IPv4Only", func(t *testing.T) {
		g := NewWithT(t)

		mockMembers := []mctypes.ClusterMember{
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node1",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("10.1.0.1:1234"),
					},
				},
			},
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node2",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("10.1.0.2:1234"),
					},
				},
			},
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node3",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("10.1.0.3:1234"),
					},
				},
			},
		}

		localhostAddress, err := app.DetermineLocalhostAddress(mockMembers)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(localhostAddress).To(Equal("127.0.0.1"))
	})

	t.Run("IPv6Only", func(t *testing.T) {
		g := NewWithT(t)

		mockMembers := []mctypes.ClusterMember{
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node1",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("[fda1:8e75:b6ef::]:1234"),
					},
				},
			},
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node2",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("[fd51:d664:aca3::]:1234"),
					},
				},
			},
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node3",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("[fda3:c11d:3cda::]:1234"),
					},
				},
			},
		}

		localhostAddress, err := app.DetermineLocalhostAddress(mockMembers)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(localhostAddress).To(Equal("[::1]"))
	})

	t.Run("IPv4_IPv6_Mixed", func(t *testing.T) {
		g := NewWithT(t)

		mockMembers := []mctypes.ClusterMember{
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node1",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("10.1.0.1:1234"),
					},
				},
			},
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node2",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("[fd51:d664:aca3::]:1234"),
					},
				},
			},
			{
				ClusterMemberLocal: mctypes.ClusterMemberLocal{
					Name: "node3",
					Address: mctypes.AddrPort{
						AddrPort: netip.MustParseAddrPort("10.1.0.3:1234"),
					},
				},
			},
		}

		localhostAddress, err := app.DetermineLocalhostAddress(mockMembers)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(localhostAddress).To(Equal("[::1]"))
	})
}
