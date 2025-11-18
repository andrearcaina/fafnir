package utils

import (
	"errors"
	"strings"
	"time"
)

func ParseTimeFlag(flag string) (time.Time, error) {
	if flag == "" {
		return time.Time{}, errors.New("time flag cannot be empty")
	}

	// try to parse as duration relative to now
	if strings.HasPrefix(flag, "now-") {
		durationStr := strings.TrimPrefix(flag, "now-")
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return time.Time{}, err
		}
		return time.Now().Add(-duration), nil
	}

	// try to parse as duration
	if duration, err := time.ParseDuration(flag); err == nil {
		return time.Now().Add(-duration), nil
	}

	// try to parse as RFC3339
	if t, err := time.Parse(time.RFC3339, flag); err == nil {
		return t, nil
	}

	// try to parse as date only (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", flag); err == nil {
		return t, nil
	}

	// try to parse as date and time (YYYY-MM-DDTHH:MM)
	if t, err := time.Parse("2006-01-02T15:04", flag); err == nil {
		return t, nil
	}

	// try to parse as date and time with seconds (YYYY-MM-DDTHH:MM:SS)
	if t, err := time.Parse("2006-01-02T15:04:05", flag); err == nil {
		return t, nil
	}

	return time.Time{}, errors.New("unable to parse time format. Supported formats: 1h, 30m, now-1h, 2006-01-02, 2006-01-02T15:04:05")
}
