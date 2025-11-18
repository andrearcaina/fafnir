package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"fafnir/tools/logctl/internal/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/joho/godotenv"
)

type Client struct {
	es     *elasticsearch.Client
	config *types.Config
}

func NewClient() (*Client, error) {
	config := loadConfig()

	cfg := elasticsearch.Config{
		Addresses: []string{config.URL},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	// test connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return &Client{
		es:     es,
		config: config,
	}, nil
}

func loadConfig() *types.Config {
	_ = godotenv.Load()

	return &types.Config{
		URL:          os.Getenv("ELASTICSEARCH_URL"),
		IndexPattern: os.Getenv("ELASTICSEARCH_INDEX_PATTERN"),
		QueryTimeout: 30 * time.Second,
		MaxResults:   10000,
	}
}

func (c *Client) QueryLogs(opts *types.QueryOptions) ([]types.LogEntry, error) {
	query := c.buildQuery(opts)

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("error marshaling query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{c.config.IndexPattern},
		Body:  strings.NewReader(string(queryJSON)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	var esRes types.ESResponse
	if err := json.NewDecoder(res.Body).Decode(&esRes); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	logs := make([]types.LogEntry, 0, len(esRes.Hits.Hits))
	for _, hit := range esRes.Hits.Hits {
		logs = append(logs, hit.Source)
	}

	return logs, nil
}

func (c *Client) buildQuery(opts *types.QueryOptions) map[string]interface{} {
	mustClauses := []map[string]interface{}{
		{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"gte": opts.Since.Format(time.RFC3339),
					"lte": opts.Until.Format(time.RFC3339),
				},
			},
		},
	}

	// add service filter
	if opts.Service != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"kubernetes.container.name": opts.Service,
			},
		})
	}

	// add request ID filter
	if opts.RequestID != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"request_id": opts.RequestID,
			},
		})
	}

	// add search filter
	if opts.Search != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"message": map[string]interface{}{
					"query":    opts.Search,
					"operator": "and",
				},
			},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"size": opts.Limit,
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{
					"order": "asc",
				},
			},
		},
	}

	return query
}
