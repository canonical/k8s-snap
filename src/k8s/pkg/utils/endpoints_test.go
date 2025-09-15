package utils

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestParseEndpoints(t *testing.T) {
	for _, tc := range []struct {
		name      string
		endpoints *v1.Endpoints
		addresses []string
	}{
		{
			name: "one",
			endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{Addresses: []v1.EndpointAddress{{IP: "1.1.1.1"}}},
				},
			},
			addresses: []string{"1.1.1.1:6443"},
		},
		{
			name: "two",
			endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{Addresses: []v1.EndpointAddress{{IP: "1.1.1.1"}, {IP: "2.2.2.2"}}},
				},
			},
			addresses: []string{"1.1.1.1:6443", "2.2.2.2:6443"},
		},
		{
			name: "IPv6",
			endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{Addresses: []v1.EndpointAddress{{IP: "fe80::e0b9:bfff:fe90:8d37"}}},
				},
			},
			addresses: []string{"[fe80::e0b9:bfff:fe90:8d37]:6443"},
		},
		{
			name: "multiple-subsets",
			endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{Addresses: []v1.EndpointAddress{{IP: "1.1.1.1"}, {IP: "2.2.2.2"}}},
					{Addresses: []v1.EndpointAddress{{IP: "3.3.3.3"}}},
				},
			},
			addresses: []string{"1.1.1.1:6443", "2.2.2.2:6443", "3.3.3.3:6443"},
		},
		{
			name: "override-port",
			endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{Addresses: []v1.EndpointAddress{{IP: "1.1.1.1"}, {IP: "2.2.2.2"}}},
					{Addresses: []v1.EndpointAddress{{IP: "3.3.3.3"}}, Ports: []v1.EndpointPort{{Port: int32(10000), Name: "https"}}},
				},
			},
			addresses: []string{"1.1.1.1:6443", "2.2.2.2:6443", "3.3.3.3:10000"},
		},
		{
			name: "sort",
			endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{Addresses: []v1.EndpointAddress{{IP: "3.3.3.3"}, {IP: "1.1.1.1"}}},
					{Addresses: []v1.EndpointAddress{{IP: "2.2.2.2"}}, Ports: []v1.EndpointPort{{Port: int32(10000), Name: "https"}}},
				},
			},
			addresses: []string{"1.1.1.1:6443", "2.2.2.2:10000", "3.3.3.3:6443"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if parsed := ParseEndpoints(tc.endpoints); !reflect.DeepEqual(parsed, tc.addresses) {
				t.Fatalf("expected addresses to be %v but they were %v instead", tc.addresses, parsed)
			}
		})
	}
}
