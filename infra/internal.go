package infra

import (
	"log"
	"net/http"
)

// StartInternal webserver for acme challenges
func StartInternal(httpHandler http.Handler) {
	srv := &http.Server{
		Addr:    ":80",
		Handler: httpHandler,
	}

	// Starting HTTP server for Let's Encrypt challenge handlers
	go func() {
		log.Printf("Starting HTTP server on %s", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("could not start HTTP server: %v", err)
		}
	}()
}
