package main

import (
	"github.com/breathbath/certs/acme"
	"github.com/breathbath/certs/domain"
	"github.com/breathbath/certs/infra"
	"github.com/breathbath/certs/redirect"
	"log"
	"net/http"
)

func main() {
	internalHandler := http.NewServeMux()

	domainPolicy := domain.NewDynamicHostPolicy()

	acmeManager := acme.NewAcmeManager(domainPolicy.AllowHost)
	internalHandler.Handle("/", acmeManager.HTTPHandler(nil))

	domainHandler := domain.NewHandler(domainPolicy.AddDomain)
	internalHandler.Handle("/add-domain", domainHandler)

	infra.StartInternal(internalHandler)

	externalHandler := http.NewServeMux()
	redirectHandler := redirect.NewRedirectHandler()
	externalHandler.Handle("/", redirectHandler)

	err := infra.StartExternal(acmeManager.GetCertificate, externalHandler)
	if err != nil {
		log.Fatalf("failed to start external HTTPS server: %v", err)
	}
}
