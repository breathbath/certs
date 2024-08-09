package redirect

import "net/http"

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func NewRedirectHandler() http.Handler {
	return &Handler{}
}
