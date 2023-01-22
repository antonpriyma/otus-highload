package cert

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

type Storage interface {
	Cert() tls.Certificate
	CA() *x509.CertPool
}

type StorageConfig struct {
	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
	CA   string `mapstructure:"ca"`
}

func (s StorageConfig) IsEmpty() bool {
	empty := StorageConfig{}
	return s == empty
}

func NewStorage(cfg StorageConfig) (Storage, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load cert %s with key %s", cfg.Cert, cfg.Key)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(cfg.CA)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read ca cert %s", cfg.CA)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return storage{
		cert: cert,
		ca:   caCertPool,
	}, nil
}

type storage struct {
	cert tls.Certificate
	ca   *x509.CertPool
}

func (s storage) Cert() tls.Certificate {
	return s.cert
}

func (s storage) CA() *x509.CertPool {
	return s.ca
}
