package pki_test

import (
	"net"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
)

func TestWorkerNodePKI_CompleteWorkerNodePKI(t *testing.T) {
	cert := `
-----BEGIN CERTIFICATE-----
MIIDHDCCAgSgAwIBAgIRAMp9M56e6mSaXAkgEkvKbuQwDQYJKoZIhvcNAQELBQAw
GDEWMBQGA1UEAxMNa3ViZXJuZXRlcy1jYTAeFw0yNDAzMTQxNTM3MjdaFw00NDAz
MTQxNTM3MjdaMBgxFjAUBgNVBAMTDWt1YmVybmV0ZXMtY2EwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQDOph9lBC0hLf2ybOcBfMQQs6AJw6/6MDe06SyY
1uGPOv0CYXsmcku5KzgCruE6Dal2vNK9WQkgTRbxjt84xjHI93/W5IGdB9ZTyGem
SSeEtXD9x71eptKCrHcwbtbbUlLwmRIuAXifVDWZqCp41HwM3HhWgH4cILywFNrp
kHfm6p7CSrFRvzldmU8DAtAUHZ4iGJoVkSKVhSY4Tj18q+5+nkPrUww1QvVJ/QXn
9pc7gwig/qrnF85GjyBCLhO7IghCeImFSRxyMDaOgfa1Fd5mF0i3I5ViHVx3wiQq
IfWzWoO76kTTwugeu6UY88MqV8Y2SPqnEL2lzYNStsJogt+nAgMBAAGjYTBfMA4G
A1UdDwEB/wQEAwIChDAdBgNVHSUEFjAUBggrBgEFBQcDAgYIKwYBBQUHAwEwDwYD
VR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUKE+9hTdaPuNtXmJzZwmM1MsG0yowDQYJ
KoZIhvcNAQELBQADggEBAIn9XRqXWqQS70fmeIB94UOxmj8TTXql0eFsw8h9NidP
aMFjZbM1ovKVhHId9n09wiTivo/S6kX7n/8IzBPiB9wmlEy6NPpLppfy7VEhUqfK
K1R9leoNoirda0FhQjXoQ1IGdHhA3Gw0woToeIfRlB+J6cMRth88/3bk/aA8ZR63
bDqf0KtLXVs90UUVehUrWtj14CzSEhsyC7hcd3FKx6yzcviiydPXqBocbdLpzv2w
Zfb6LUptXDMSQxlU+meP6PjZtSxR59HivhrtSqkZd14bW01Pi9zGHvccgOsGI0hi
8+MoDI4x8YcQdkn4uA3wT88spYHOLAsUXLxK9tPlzr8=
-----END CERTIFICATE-----`

	key := `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAzqYfZQQtIS39smznAXzEELOgCcOv+jA3tOksmNbhjzr9AmF7
JnJLuSs4Aq7hOg2pdrzSvVkJIE0W8Y7fOMYxyPd/1uSBnQfWU8hnpkknhLVw/ce9
XqbSgqx3MG7W21JS8JkSLgF4n1Q1magqeNR8DNx4VoB+HCC8sBTa6ZB35uqewkqx
Ub85XZlPAwLQFB2eIhiaFZEilYUmOE49fKvufp5D61MMNUL1Sf0F5/aXO4MIoP6q
5xfORo8gQi4TuyIIQniJhUkccjA2joH2tRXeZhdItyOVYh1cd8IkKiH1s1qDu+pE
08LoHrulGPPDKlfGNkj6pxC9pc2DUrbCaILfpwIDAQABAoIBAFe1Wo3df+odQxh/
8GxJME6Gbt62F/LwlDRM44jba1EHkGt6RHLFAC7PkS5SW3XwZoTnD+sd5ym2jo5o
PYYzWN4bbj8fLYQg128oGBYT5poFCLguFsodtCuSV+ROpxLfliRYU8cDCNdXPojB
P4WZai1rRggw8VWu72cs8t0/XCS9nNwzO8er/PEQiMM3RQtAYoANMc/l4GiHW7jk
7+ASZK2WWH+zEPVvtmhNmLrPHtl0AvEoj/PISNoHcUqJ7+ab/YdrRx4Kie8w2rF2
ieMqSoXb1X5dHoPR5h0maP3XVe4JuhrrukOQgTedkban/cfwZgbvkrNR3t8oMKKa
EH8i8gECgYEA3seCu4FNmavhXxjmw6Zk2CMT5WKcTOmc35jQxOFaEUjsn8+4ONXR
O/JkQfJMG78XX7K3zIaLsV/xQS2JDfXPr1eZGlY1d3GuiZbkMw2rFV50PvssSr17
OgCLiAj3RbWDCZbCJ5t/x8azEBq8v3cJM81CvACgO0WLvkYovl9tsOECgYEA7XbX
WS77jsmF7YYfzRYhMl7iFYt+zumbz+84xTS8pxgkS6Bk/54rAcpq4vPmeQ0SFuGn
h8CN6TGQ4e8Qt0ea57jMmW74Lpgm4Gk1/lGahELZDJXgfvuoAcjAA67lJi/DheH0
mfMsxgC5M0BjvRfS+K8Qwh7sn6uDFBkxlfPYuYcCgYBJ52uyIlII8aEhOBSNwSxh
GznlddIeHb2h24MeXRfQ9h0xYupdSGlR9rZVvjiLV9g8MgCRQ+0hmY9iLOXzkKEm
LOwodYLlLfxVvo3TdexUeXIc1pw56yPu+PFQ3pCROobO7olYNFiugHc0l3oYFjgi
TCygS6DcKNUT+RhZFzU/YQKBgQCwagmyh+T7P1vwCiS2CCrBcRwlRWz/6y2GXQKf
/33n5VeRl6dw/+CTg/3Efc5LQBqgRSRhBfxnshsgvqp8fwXmALR/iKF4fDDlp0Ql
nBpfCAqX/wC5VdyK9skv807p/7ISVLuTY8VvlDoCiWOPp5NkjSq2DKNeO901oUHl
VTM9IQKBgQChWOTuFetvVLgMa7kc0y6TYkCRsdb5XxsGkDwQv5EyctG+w5EXB0LM
li9E9xBQfc5nb88jQ4hf+9wjm0Q15LsocSNxtbUN+F5T4cskQAxFB4/djpBSbieu
IAoJNLRY/jIMGkQkzRdS1oXGXZqsAW9ndS5+N6uF7+SaXBO7E5yFKA==
-----END RSA PRIVATE KEY-----`
	tests := []struct {
		name          string
		pki           *pki.ControlPlanePKI
		hostname      string
		nodeIP        net.IP
		bits          int
		expectedError bool
	}{
		{
			name:          "CompleteWorkerNodePKI with missing CA certificate",
			pki:           &pki.ControlPlanePKI{},
			expectedError: true,
		},
		{
			name:          "CompleteWorkerNodePKI with CA certificate but without CA key",
			pki:           &pki.ControlPlanePKI{CACert: cert},
			expectedError: false,
		},
		{
			name:          "CompleteWorkerNodePKI with CA certificate and successful certificate generation",
			pki:           &pki.ControlPlanePKI{CACert: cert, CAKey: key},
			hostname:      "worker-node-1",
			nodeIP:        net.ParseIP("10.152.183.1"),
			bits:          2048,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pki, err := tt.pki.CompleteWorkerNodePKI(tt.hostname, tt.nodeIP, tt.bits)

			if (err != nil) != tt.expectedError {
				t.Errorf("Unexpected error status. Expected error: %v, got error: %v", tt.expectedError, err)
			}

			if !tt.expectedError {
				if pki.CACert == "" {
					t.Error("Missing certificate details in completed worker node PKI")
				}
			}
		})
	}
}
