package main

import (
	"github.com/breathbath/certs/acme"
	"github.com/breathbath/certs/domain"
	"github.com/breathbath/certs/infra"
	"github.com/breathbath/certs/proxy"
	"log"
	"net/http"
)

func main() {
	internalRoutes := http.NewServeMux()
	externalRoutes := http.NewServeMux() //this is an internet facing webserver with https

	domainStorage, err := domain.NewStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer domainStorage.Close()

	domainPolicy := domain.NewDynamicHostPolicy(domainStorage)

	acmeManager := acme.NewAcmeManager(domainPolicy.AllowHost)
	internalRoutes.Handle("/", acmeManager.HTTPHandler(nil))
	infra.StartInternal(internalRoutes)

	domainHandler := domain.NewHandler(domainStorage)
	domainHandler.RegisterRoutes(externalRoutes)

	reverseProxyHandler := proxy.NewReverseProxyHandler(domainStorage)
	externalRoutes.Handle("/", reverseProxyHandler)

	err = infra.StartExternal(acmeManager.GetCertificate, externalRoutes)
	if err != nil {
		log.Fatalf("failed to start external HTTPS server: %v", err)
	}
}
