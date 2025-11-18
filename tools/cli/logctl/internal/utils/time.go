package utils

import (
	"errors"
	"strings"
	"time"
)

// ParseTimeFlag parses various time formats
func ParseTimeFlag(flag string) (time.Time, error) {
	if flag == "" {
		return time.Time{}, errors.New("time flag cannot be empty")
	}

	// Try to parse as duration relative to now
	if strings.HasPrefix(flag, "now-") {
		durationStr := strings.TrimPrefix(flag, "now-")
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return time.Time{}, err
		}
		return time.Now().Add(-duration), nil
	}

	// Try to parse as duration
	if duration, err := time.ParseDuration(flag); err == nil {
		return time.Now().Add(-duration), nil
	}

	// Try to parse as RFC3339
	if t, err := time.Parse(time.RFC3339, flag); err == nil {
		return t, nil
	}

	// Try to parse as date only (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", flag); err == nil {
		return t, nil
	}

	// Try to parse as date and time (YYYY-MM-DDTHH:MM)
	if t, err := time.Parse("2006-01-02T15:04", flag); err == nil {
		return t, nil
	}

	// Try to parse as date and time with seconds (YYYY-MM-DDTHH:MM:SS)
	if t, err := time.Parse("2006-01-02T15:04:05", flag); err == nil {
		return t, nil
	}

	return time.Time{}, errors.New("unable to parse time format. Supported formats: 1h, 30m, now-1h, 2006-01-02, 2006-01-02T15:04:05")
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return formatTime(hours, "h", minutes, "m", seconds, "s")
	}
	if minutes > 0 {
		return formatTime(minutes, "m", seconds, "s")
	}
	return formatTime(seconds, "s")
}

func formatTime(values ...interface{}) string {
	var parts []string
	for i := 0; i < len(values); i += 2 {
		val := values[i].(int)
		unit := values[i+1].(string)
		if val > 0 {
			parts = append(parts, formatPart(val, unit))
		}
	}
	return strings.Join(parts, " ")
}

func formatPart(value int, unit string) string {
	if value == 1 {
		return "1" + unit
	}
	return string(rune(value)) + unit
}

// TimePresets returns common time presets
func TimePresets() map[string]time.Duration {
	return map[string]time.Duration{
		"Last 5 minutes":  5 * time.Minute,
		"Last 15 minutes": 15 * time.Minute,
		"Last 30 minutes": 30 * time.Minute,
		"Last 1 hour":     1 * time.Hour,
		"Last 3 hours":    3 * time.Hour,
		"Last 6 hours":    6 * time.Hour,
		"Last 12 hours":   12 * time.Hour,
		"Last 24 hours":   24 * time.Hour,
		"Last 7 days":     7 * 24 * time.Hour,
	}
}
