package types

import "time"

type QueryOptions struct {
	Service   string
	Since     time.Time
	Until     time.Time
	Search    string
	Limit     int
	RequestID string
}

type LogEntry struct {
	Timestamp  time.Time  `json:"@timestamp"`
	Message    string     `json:"message"`
	Kubernetes Kubernetes `json:"kubernetes,omitempty"`
	RequestID  string     `json:"request_id,omitempty"`
	Error      string     `json:"error,omitempty"`
	UserID     string     `json:"user_id,omitempty"`
	TraceID    string     `json:"trace_id,omitempty"`
	Method     string     `json:"method,omitempty"`
	Path       string     `json:"path,omitempty"`
	Status     int        `json:"status,omitempty"`
	Duration   float64    `json:"duration_ms,omitempty"`
}

type Kubernetes struct {
	Container Container `json:"container,omitempty"`
}

type Container struct {
	Name string `json:"name,omitempty"`
}

type ESResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source LogEntry `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type Config struct {
	URL          string
	IndexPattern string
	MaxResults   int
	QueryTimeout time.Duration
}
