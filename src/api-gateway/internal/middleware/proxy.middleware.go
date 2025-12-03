package middleware

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// this transport configuration is optimized for high concurrency and low latency
var customProxyTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          1000, // this allows many idle connections
	MaxIdleConnsPerHost:   1000, // this allows many concurrent connections to the same host
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func ReverseProxy(targetURL string) http.Handler {
	uri, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(uri)
	proxy.Transport = customProxyTransport

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = uri.Host
		proxy.ServeHTTP(w, r)
	})
}
