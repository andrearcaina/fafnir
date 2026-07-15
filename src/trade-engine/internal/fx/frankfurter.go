package fx

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"resty.dev/v3"
)

type cachedRate struct {
	rate      float64
	expiresAt time.Time
}

type Frankfurter struct {
	client *resty.Client
	ttl    time.Duration

	mu    sync.RWMutex
	cache map[string]cachedRate
}

func NewFrankfurter(baseURL string, timeout time.Duration, ttl time.Duration) *Frankfurter {
	client := resty.New().
		SetBaseURL(strings.TrimRight(baseURL, "/")).
		SetTimeout(timeout)

	return &Frankfurter{
		client: client,
		ttl:    ttl,
		cache:  make(map[string]cachedRate),
	}
}

func (f *Frankfurter) Close() error {
	return f.client.Close()
}

func (f *Frankfurter) Rate(ctx context.Context, from string, to string) (float64, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))
	if from == "" || to == "" {
		return 0, fmt.Errorf("currency codes must not be empty")
	}
	if from == to {
		return 1, nil
	}

	key := from + ":" + to
	if rate, ok := f.cached(key); ok {
		return rate, nil
	}

	var payload struct {
		Rate float64 `json:"rate"`
	}

	endpoint := fmt.Sprintf("/v2/rate/%s/%s", url.PathEscape(from), url.PathEscape(to))
	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&payload).
		Get(endpoint)
	if err != nil {
		return 0, fmt.Errorf("fetch %s/%s FX rate: %w", from, to, err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("fetch %s/%s FX rate: unexpected status %d", from, to, resp.StatusCode())
	}
	if payload.Rate <= 0 {
		return 0, fmt.Errorf("FX provider returned an invalid %s/%s rate", from, to)
	}

	f.mu.Lock()
	f.cache[key] = cachedRate{rate: payload.Rate, expiresAt: time.Now().Add(f.ttl)}
	f.mu.Unlock()

	return payload.Rate, nil
}

func (f *Frankfurter) cached(key string) (float64, bool) {
	f.mu.RLock()
	entry, ok := f.cache[key]
	f.mu.RUnlock()

	return entry.rate, ok && time.Now().Before(entry.expiresAt)
}
