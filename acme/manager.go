package acme

import (
	"golang.org/x/crypto/acme/autocert"
	"log"
	"os"
)

func NewAcmeManager(hp autocert.HostPolicy) *autocert.Manager {
	cacheDir := ".certs"
	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		log.Fatalf("could not create certs directory: %v", err)
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache(cacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hp,
	}

	return m
}
