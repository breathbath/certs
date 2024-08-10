package proxy

import (
	"github.com/breathbath/certs/utils"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
)

type TargetProvider interface {
	GetTarget(host string) string
}

type ReverseProxyHandler struct {
	targetProvider TargetProvider
	targetsCache   sync.Map
}

func NewReverseProxyHandler(targetProvider TargetProvider) *ReverseProxyHandler {
	return &ReverseProxyHandler{
		targetProvider: targetProvider,
		targetsCache:   sync.Map{},
	}
}

func (h *ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r, true)

	target := h.targetProvider.GetTarget(r.Host)
	if target == "" {
		log.Printf("failed to find a supported target for host %s\n", r.Host)

		http.Error(w, "Unsupported host: "+r.Host, http.StatusNotFound)
		return
	}

	log.Printf("host: %s, target: %s\n", r.Host, target)

	cachedProxyI, ok := h.targetsCache.Load(target)
	if ok {
		log.Println("loaded proxy from cache")

		cachedProxy, ok := cachedProxyI.(*httputil.ReverseProxy)
		if ok {
			cachedProxy.ServeHTTP(w, r)
			return
		}
	}

	log.Printf("created proxy for target %s\n", target)

	rp, err := NewReverseProxy(target)
	if err != nil {
		log.Printf("invalid target: %s: %v\n", target, err)
		http.Error(w, "Wrong target configuration", http.StatusUnprocessableEntity)
		return
	}

	h.targetsCache.Store(target, rp)
	rp.ServeHTTP(w, r)
}

func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	proxyTargetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Ivalid target url %s: %v", target, err)
	}

	rp := httputil.NewSingleHostReverseProxy(proxyTargetURL)
	originalDirector := rp.Director
	rp.Director = func(req *http.Request) {
		req.Header.Set("X-Original-Host", req.Host)
		// Preserve the original X-Forwarded-For header if it exists, and add the client's IP
		clientIP := req.RemoteAddr
		if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			req.Header.Set("X-Forwarded-For", forwardedFor+", "+clientIP)
		} else {
			req.Header.Set("X-Forwarded-For", clientIP)
		}

		req.Header.Set("X-Real-IP", clientIP)

		originalDirector(req)

		utils.LogRequest(req, false)
	}

	errorLog := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	rp.ErrorLog = errorLog

	return rp, nil
}
