package mailer_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
)

var trustedCertificate, untrustedCertificate tls.Certificate

// TestMain confines test certificate trust to this test process. Production
// still uses its normal certificate verification and system trust settings.
func TestMain(m *testing.M) {
	var err error
	trustedCertificate, err = newSMTPTestCertificate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	untrustedCertificate, err = newSMTPTestCertificate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	roots := x509.NewCertPool()
	roots.AddCert(trustedCertificate.Leaf)
	debug := os.Getenv("GODEBUG")
	if debug != "" {
		debug += ","
	}
	if err = os.Setenv("GODEBUG", debug+"x509usefallbackroots=1"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	x509.SetFallbackRoots(roots)
	os.Exit(m.Run())
}

func newSMTPTestCertificate() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "SMTP test server"},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true, IsCA: true,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}, nil
}
