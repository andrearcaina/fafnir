package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ReverseProxyMiddleware(targetURL string) http.Handler {
	uri, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(uri)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// optionally modify headers here
		r.Host = uri.Host
		proxy.ServeHTTP(w, r)
	})
}
