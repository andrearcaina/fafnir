package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ReverseProxy(targetURL string) http.Handler {
	uri, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(uri)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = uri.Host
		proxy.ServeHTTP(w, r)
	})
}
