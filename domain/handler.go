package domain

import (
	"log"
	"net/http"
)

type Handler struct {
	AddDomain func(domain string)
}

func NewHandler(addDomain func(domain string)) *Handler {
	return &Handler{
		AddDomain: addDomain,
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Unsupported method. Use POST.", http.StatusMethodNotAllowed)
		return
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(rw, "Domain parameter is required", http.StatusBadRequest)
		return
	}

	h.AddDomain(domain)

	log.Printf("Domain added: %s", domain)

	rw.Write([]byte("Domain added"))
}
