package redirect

import (
	"net"
	"net/http"
)

type TargetProvider interface {
	GetRedirectTarget(host string) string
}

type Handler struct {
	targetProvider TargetProvider
}

func (h *Handler) RegisterRoutes(routes *http.ServeMux) {
	routes.HandleFunc("/", h.Redirect)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	host := h.getHostname(r)
	target := h.targetProvider.GetRedirectTarget(host)
	if target == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func (h *Handler) getHostname(r *http.Request) string {
	host := r.Host

	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}

	return hostname
}

func NewRedirectHandler(targetProvider TargetProvider) *Handler {
	return &Handler{
		targetProvider: targetProvider,
	}
}
