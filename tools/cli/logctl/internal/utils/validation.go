package utils

import (
	"errors"
	"strings"
)

func ValidServicesList() []string {
	return []string{"auth-service", "user-service", "security-service", "stock-service", "api-gateway"}
}

func ValidateService(service string) error {
	if service == "" {
		return errors.New("service name cannot be empty")
	}

	for _, valid := range ValidServicesList() {
		if service == valid {
			return nil
		}
	}

	return errors.New("invalid service name. Valid services: " + strings.Join(ValidServicesList(), ", "))
}

func ValidateLimit(limit int) error {
	if limit <= 0 {
		return errors.New("limit must be greater than 0")
	}
	if limit > 10000 {
		return errors.New("limit cannot exceed 10000")
	}
	return nil
}

func ValidateInterval(interval int) error {
	if interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	if interval > 60 {
		return errors.New("interval cannot exceed 60 seconds")
	}
	return nil
}

func NormalizeService(service string) string {
	return strings.ToLower(strings.TrimSpace(service))
}
