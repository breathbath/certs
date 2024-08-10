package infra

import (
	"crypto/tls"
	"log"
	"net/http"
)

type CertificateProvider func(info *tls.ClientHelloInfo) (*tls.Certificate, error)

func StartExternal(certProvider CertificateProvider, h http.Handler) error {
	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: certProvider,
		},
		Handler: h,
	}

	log.Printf("Starting HTTPS server on %s", server.Addr)
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		return err
	}

	return nil
}
