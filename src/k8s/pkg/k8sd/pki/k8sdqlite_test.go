package pki

import (
	"net"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestK8sDqlitePKI_CompleteCertificates(t *testing.T) {
	notBefore := time.Now()

	cert := `
-----BEGIN CERTIFICATE-----
MIIDtTCCAp2gAwIBAgIQOPOTOjxvIVlC5ev8EzrnITANBgkqhkiG9w0BAQsFADAY
MRYwFAYDVQQDEw1rdWJlcm5ldGVzLWNhMB4XDTI0MDIwODAyNDYyOVoXDTM0MDIw
ODAyNDYyOVowGTEXMBUGA1UEAxMOa3ViZS1hcGlzZXJ2ZXIwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQCum3KkohfK+E4KCpauilnlxm0e6y+jzyOaRCHx
P/3iLqN5zN+s2SV+GJNNcT3vSVZ1YhcJKWNrs7QxK2qcq9OhHncmp9Vqu5BV9O+e
ys4bBlf08lHH0//wrAwXy71ueWXN2uWyFg4i2VSirbRxpXGIR751i4qVtutbSOPy
3Jjf07upq3zAMyvTx1YTZcwduwW2vrU1f48IZOTueS1eOz0YjCkWLueD2uhLLgRA
mcxq33pwTM9P0MaZGrrM2GeA+1Hyss5WtoEMkR6TPUWQmYcKFEZui9/JpLfbM8yu
6h6Ta7GeSccjtclHSGp9fge0IXErhYSmLNoQ7JP8fQeg0DpTAgMBAAGjgfkwgfYw
DgYDVR0PAQH/BAQDAgSwMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATAM
BgNVHRMBAf8EAjAAMB8GA1UdIwQYMBaAFJjD6HMwGRJQMOzNm919/ZaqdcUwMIGV
BgNVHREEgY0wgYqCCmt1YmVybmV0ZXOCEmt1YmVybmV0ZXMuZGVmYXVsdIIWa3Vi
ZXJuZXRlcy5kZWZhdWx0LnN2Y4Iea3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVz
dGVygiRrdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWyHBAqYtwGH
BH8AAAEwDQYJKoZIhvcNAQELBQADggEBADPWn//rPb0SmZ49WhIa6wc39Ryl7eGo
Q2H+RY9BMah/xge6fLgeTvFe+H6Vol9BVqm5XgD0BuD5yzKYI2aDq8Ikm4EMOxPl
7Gs9cqWMMF7Iiw+rYJY4vwzm+5kSCg6oxBx8GLYYkDpbFe8UAWKf/9QTghtoBEEw
JVBDECnQwJU4tb9ANmPbgxmCYLZjx2vmXQRlXpe6QS9nPmMSS9KkJMyLEEpgzIIA
aSprnA8WIeSaO/5wLMYS1lUWWzegz2LnKuJ5C5Q+XYkwIY/vFH7OSTnmvt+rHwhh
4Oj+ScJ0RKnGGcXQnctSvMogDoucw7Y2RjxKcJV8fEKV5ZIeTz0U+nE=
-----END CERTIFICATE-----`

	key := `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEArAFyCH36MjE3xAXyrjhZ0B6dtH5VLZvBdflxXFdasy9ecsQk
IjKGSB3QqpldWqiRhgFnkaKnvoizpKTKOf5AYJwtdIAeR43SfIIUcyOeIY7m7Dz/
Jd8Jyq4t9rP2kI5GNYW/OVdDSatPr2C+UBQXXriNDvYcxdnWXoJ2LKUL/o7rZdJ0
Be3uL9QbH43WOq2GrJYCgJNgNfKWMAK8jhpNEpVVsvPh5mS47443LZbkTyq3StL3
849wH+8bK6YVCpdpBNorjKGblkzPD+iw5um6elsWgFq7/8l8LN0I++NM2aAMB1Ao
9Au1EX7zRzFdrM52YgfBMDyouRKLUjDWCwpxDQIDAQABAoIBADHJBVGR7Q4EEukI
87Ibm1tS0UDB5DOcRoW4Gmio3BbLGiJLxU2kpBtRjekjFNM9wUkxNOIBW14ZwS1h
iSr5/XY5Hir/PkRlt0vUdsjQwV9jNlGgYhV7FiF1AtbKRg6XL5kkSjH1oQM8s4bG
kK8q8Yy4DBQNhkx5/cNDLaNEblFSYOx7fsdfqVH4WLoXguaLgXehwWyK08NkpkaF
EC47S/X/nqPn4JOwQpVGVBQArun4ATNKLokKZqAIypjtIQl+7V1YEJAIJqbUtLDs
IuksAXiVISbXdhZm/gUH6Ok9RCcKpJ9VzqraKNA3h+LNgcbJKw+7MijS7Yv3/7/z
jO90YeECgYEA4J5ufiMa/tNzc4b3TcrNnxcJ1cdDfkxqtq0zv2ByDu46CEHbNI0H
TCujL4MEwPG1N96apA9IpZcY1rDirIbcFYAMgsYti7cRAqLv5bj8eDnYn174qHHL
jij7R1hGTcVWBkwzyCaPgCjVaBcW2KwOuEmFxb/SFpfKqY4CEl05WVkCgYEAxAlI
ykUKLVMY1aoaK48OxmjtgUU7azpgwOaCnu8SOM6mrk6NzMJq7i6P9MgFjfV2UZZc
qRZK3xGw+Gh3gkhunllJk/2S3QwkeI4NWZGDQT7QhzDQXT43vVlunKHDD/+rpC8r
CNC2hrAOgPCulipe2HDR9Y4C3WWdnEOeWS8uqtUCgYEAoEGpB6m4SvNGPbifnQsC
pWzkgXfHucZ/pJHyh6oh9nEVSmriIJ42BKxlozJRI+/PoWra3g5hgHNLL3HIZ9tY
DqbrRipquHIGWuExU68lwglTenFh65w05NpsXTyn/Di85YVctIJ+g6uehsNic3he
kDE0liADnkbyOwKsi7mjfxECgYEAmaK1G2DUQwVW900iyXSKndDqIl/B252a6lM9
l5XB8Cd01jLWSt0rtJNlWu/P+pufKP3wjMvdzcktquEkmERv/UX4tjUK/pZfluOt
br7t4Rp7jxgglJMIWCtY1wSnvUggmsIktfnsss4T79Ww3htCzdpNkmbDtAPJbAhK
d5bUuikCgYEAp+mdk4PZ2P2KW5ygXxyAU6Xmc35gLHWrKZqLY9sAnp0mA/G8c1FL
TXZaEqvvBYG4LEl1QhXukLHE2tP0Azj6+Tg+dhwiPua4OcsZpcPXIBRscris8OgL
m5cIDhPBuZSCs7ZnhWCHF0WMztl6fqNVp2GuFGbDM+LjAZT2YOdP0Ts=
-----END RSA PRIVATE KEY-----
`

	tests := []struct {
		name          string
		pki           *K8sDqlitePKI
		expectedError bool
	}{
		{
			name:          "CompleteCertificates with missing K8sDqliteCert and K8sDqliteKey self-signing not allowed",
			pki:           &K8sDqlitePKI{},
			expectedError: true,
		},
		{
			name: "CompleteCertificates with K8sDqliteCert but missing K8sDqliteKey",
			pki: &K8sDqlitePKI{
				K8sDqliteCert: cert,
			},
			expectedError: true,
		},
		{
			name: "CompleteCertificates with missing K8sDqliteCert but K8sDqliteKey",
			pki: &K8sDqlitePKI{
				K8sDqliteKey: key,
			},
			expectedError: true,
		},
		{
			name: "CompleteCertificates with K8sDqliteCert and K8sDqliteKey",
			pki: &K8sDqlitePKI{
				K8sDqliteCert: cert,
				K8sDqliteKey:  key,
			},
			expectedError: false,
		},
		{
			name: "CompleteCertificates with self-signed CA and successful certificate generation",
			pki: &K8sDqlitePKI{
				allowSelfSignedCA: true,
				hostname:          "localhost",
				notBefore:         notBefore,
				notAfter:          notBefore.AddDate(1, 0, 0),
				ipSANs:            []net.IP{net.ParseIP("127.0.0.1")},
				dnsSANs:           []string{"localhost"},
			},
			expectedError: false,
		},
		{
			name: "CompleteCertificates with self-signed CA allowed",
			pki: &K8sDqlitePKI{
				allowSelfSignedCA: true,
				hostname:          "localhost",
				notBefore:         notBefore,
				notAfter:          notBefore.AddDate(1, 0, 0),
				ipSANs:            []net.IP{},
				dnsSANs:           []string{"localhost"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := tt.pki.CompleteCertificates()

			if tt.expectedError {
				g.Expect(err).To(HaveOccurred(), "Expected error but none occurred")
			} else {
				g.Expect(err).NotTo(HaveOccurred(), "Unexpected error occurred")
			}
		})
	}
}

func TestNewK8sDqlitePKI(t *testing.T) {
	notBefore := time.Now()
	tests := []struct {
		name        string
		opts        K8sDqlitePKIOpts
		expectedPki *K8sDqlitePKI
	}{
		{
			name: "NewK8sDqlitePKI with default values",
			opts: K8sDqlitePKIOpts{
				Hostname:  "localhost",
				NotBefore: notBefore,
			},
			expectedPki: &K8sDqlitePKI{
				hostname:  "localhost",
				notBefore: notBefore,
				notAfter:  notBefore.AddDate(1, 0, 0),
			},
		},
		{
			name: "NewK8sDqlitePKI with custom values",
			opts: K8sDqlitePKIOpts{
				Hostname:          "localhost",
				DNSSANs:           []string{"localhost"},
				IPSANs:            []net.IP{net.ParseIP("127.0.0.1")},
				NotBefore:         notBefore,
				NotAfter:          notBefore.AddDate(2, 0, 0),
				AllowSelfSignedCA: true,
				Datastore:         "k8s-dqlite",
			},
			expectedPki: &K8sDqlitePKI{
				hostname:          "localhost",
				ipSANs:            []net.IP{net.ParseIP("127.0.0.1")},
				dnsSANs:           []string{"localhost"},
				notBefore:         notBefore,
				notAfter:          notBefore.AddDate(2, 0, 0),
				allowSelfSignedCA: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			pki := NewK8sDqlitePKI(tt.opts)
			g.Expect(pki).To(Equal(tt.expectedPki), "Unexpected K8sDqlitePKI")
		})
	}
}
