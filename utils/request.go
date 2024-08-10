package utils

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func LogRequest(r *http.Request, withBody bool) {
	reqDump, err := httputil.DumpRequest(r, withBody)
	if err != nil {
		log.Printf("failed to dump request: %v\n", err)
	}

	log.Printf("Request: \nRequest URL: %s\n%s\n", r.URL.String(), string(reqDump))
}
