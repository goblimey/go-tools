package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"
)

// TLS LINT
type TLS struct {
	Country    []string "GB"
	Org        []string ""
	CommonName string   "*.domain.com"
}

// Config lint
type Config struct {
	Remotehost string
	Localhost  string
	Localport  int
	TLS        *TLS
	CertFile   string ""
}

var config Config
var ids = 0

func genCert() ([]byte, *rsa.PrivateKey) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Country:      config.TLS.Country,
			Organization: config.TLS.Org,
			CommonName:   config.TLS.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey
	caB, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		fmt.Println("create ca failed", err)
	}
	return caB, priv
}

func tlsListen() (conn net.Listener, err error) {
	var cert tls.Certificate

	if config.CertFile != "" {
		cert, _ = tls.LoadX509KeyPair(fmt.Sprint(config.CertFile, ".pem"), fmt.Sprint(config.CertFile, ".key"))
	} else {
		fmt.Println("[*] Generating cert")
		caB, priv := genCert()
		cert = tls.Certificate{
			Certificate: [][]byte{caB},
			PrivateKey:  priv,
		}
	}

	conf := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	conf.Rand = rand.Reader

	conn, err = tls.Listen("tcp", fmt.Sprint(config.Localhost, ":", config.Localport), &conf)
	return
}
