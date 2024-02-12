package pki_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	. "github.com/onsi/gomega"
)

func TestControlPlaneCertificates(t *testing.T) {
	c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:          "h1",
		Years:             10,
		AllowSelfSignedCA: true,
	})

	g := NewWithT(t)

	g.Expect(c.CompleteCertificates()).To(BeNil())
	g.Expect(c.CompleteCertificates()).To(BeNil())

	t.Run("MissingCAKey", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname: "h1",
			Years:    10,
		})
		c.CACert = `
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

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).ToNot(BeNil())
	})
}
