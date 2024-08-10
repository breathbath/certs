package domain

import (
	"crypto/subtle"
	"fmt"
	"github.com/breathbath/certs/utils"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
)

type Handler struct {
	storage Storage
	authKey string
}

func NewHandler(storage Storage) *Handler {
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

func (h *Handler) checkAccess(r *http.Request) int {
	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain != "" {
		if r.Host != appDomain {
			log.Printf("%s is not allowed to access this endpoint\n", r.Host)
			return http.StatusBadRequest
		}
	}
	authKeyProvided := r.Header.Get("X-Auth-Key")
	result := subtle.ConstantTimeCompare([]byte(authKeyProvided), []byte(h.authKey))

	if result != 1 {
		log.Println("Invalid authorisation data")

		return http.StatusUnauthorized
	}

	return 0
}

func (h *Handler) AddDomain(rw http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r, false)

	accessCode := h.checkAccess(r)

	if accessCode > 0 {
		http.Error(rw, "Invalid request", accessCode)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("invalid http method: %s\n", r.Method)
		http.Error(rw, "Unsupported method. Use POST.", http.StatusMethodNotAllowed)
		return
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		log.Println("empty domain parameter in request")
		http.Error(rw, "Domain parameter is required", http.StatusBadRequest)
		return
	}
	target := r.URL.Query().Get("target")
	if target == "" {
		log.Println("empty target parameter in request")
		http.Error(rw, "Target parameter is required", http.StatusBadRequest)
		return
	}

	h.storage.Add(domain, target)

	log.Printf("Domain added: %s, target: %s\n", domain, target)

	rw.Write([]byte("Domain added"))
}

func (h *Handler) RemoveDomain(rw http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r, false)

	accessCode := h.checkAccess(r)
	if accessCode > 0 {
		http.Error(rw, "Invalid request", accessCode)
		return
	}

	if r.Method != http.MethodDelete {
		log.Printf("invalid http method: %s\n", r.Method)
		http.Error(rw, "Unsupported method. Use DELETE.", http.StatusMethodNotAllowed)
		return
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		log.Println("empty domain parameter in request")
		http.Error(rw, "Domain parameter is required", http.StatusBadRequest)
		return
	}

	h.storage.Remove(domain)

	log.Printf("Domain removed: %s\n", domain)

	rw.Write([]byte("Domain removed"))
}
