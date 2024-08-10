package domain

import (
	"crypto/subtle"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"net/http"
	"os"
)

type Handler struct {
	storage *Storage
	authKey string
}

func NewHandler(storage *Storage) *Handler {
	authKey := os.Getenv("AUTH_KEY")
	if authKey == "" {
		newUUID := uuid.New()

		authKey = newUUID.String()
		fmt.Printf("AUTH_KEY env variable is empty, generated a new random auth key: %s\n", authKey)
	}

	return &Handler{
		storage: storage,
		authKey: authKey,
	}
}

func (h *Handler) RegisterRoutes(routes *http.ServeMux) {
	routes.HandleFunc("/add-domain", h.AddDomain)
	routes.HandleFunc("/remove-domain", h.RemoveDomain)
}

func (h *Handler) checkAccess(r *http.Request) bool {
	if h.isLoopbackRequest(r) {
		return true
	}

	authKeyProvided := r.Header.Get("X-Auth-Key")
	result := subtle.ConstantTimeCompare([]byte(authKeyProvided), []byte(h.authKey))

	return result == 1
}

func (h *Handler) isLoopbackRequest(r *http.Request) bool {
	// Extract the remote address
	remoteAddr := r.RemoteAddr

	// Split the remote address to isolate the IP part (exclude the port)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	// Parse the IP address
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Check if the IP address is a loopback address
	return ip.IsLoopback()
}

func (h *Handler) AddDomain(rw http.ResponseWriter, r *http.Request) {
	if !h.checkAccess(r) {
		http.Error(rw, "Invalid credentials provided", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(rw, "Unsupported method. Use POST.", http.StatusMethodNotAllowed)
		return
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(rw, "Domain parameter is required", http.StatusBadRequest)
		return
	}
	redirectTarget := r.URL.Query().Get("redirect")
	if redirectTarget == "" {
		http.Error(rw, "Redirect target parameter is required", http.StatusBadRequest)
		return
	}

	h.storage.Add(domain, redirectTarget)

	log.Printf("Domain added: %s, target: %s", domain, redirectTarget)

	rw.Write([]byte("Domain added"))
}

func (h *Handler) RemoveDomain(rw http.ResponseWriter, r *http.Request) {
	if !h.checkAccess(r) {
		http.Error(rw, "Invalid credentials provided", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(rw, "Unsupported method. Use DELETE.", http.StatusMethodNotAllowed)
		return
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(rw, "Domain parameter is required", http.StatusBadRequest)
		return
	}

	h.storage.Remove(domain)

	log.Printf("Domain removed: %s", domain)

	rw.Write([]byte("Domain removed"))
}
