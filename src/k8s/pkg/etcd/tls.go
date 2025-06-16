package etcd

import (
	"crypto/tls"

	"go.etcd.io/etcd/client/pkg/v3/transport"
)

type tlsConfig struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

func (c tlsConfig) ClientConfig() (*tls.Config, error) {
	if c.CertFile == "" && c.KeyFile == "" && c.CAFile == "" {
		return nil, nil
	}

	info := &transport.TLSInfo{
		CertFile:      c.CertFile,
		KeyFile:       c.KeyFile,
		TrustedCAFile: c.CAFile,
	}
	tlsConfig, err := info.ClientConfig()
	if err != nil {
		return nil, err
	}

	return tlsConfig, nil
}
