// Package configs provides functionality for loading and parsing notification configurations.
package configs

import (
	"encoding/json"
	"os"
	"time"
)

// Unit represents the unit of measurement for rate limits.
type Unit string

// LimitConfigMap is a map of limit configurations, where the key is the limit type.
type LimitConfigMap map[string]*LimitConfig

// Get retrieves the limit configuration for the specified type from the map.
func (lcm LimitConfigMap) Get(typ string) *LimitConfig {
	return lcm[typ]
}

// NotificationConfig represents the configuration for the notification service.
type NotificationConfig struct {
	Gateway   string           `json:"gateway"`
	RateLimit *RateLimitConfig `json:"rate_limit"`
}

// LimitConfig represents the configuration for a rate limit.
type LimitConfig struct {
	Type    string `json:"type"`
	Limit   int64  `json:"limit"`
	WSizeMs int64  `json:"window_size_ms"`
}

// RateLimitConfig represents the configuration for rate limits.
type RateLimitConfig struct {
	Type   string         `json:"type"`
	Limits []*LimitConfig `json:"limits"`
}

// NotificationService represents the notification service with its configurations.
type NotificationService struct {
	GatewayType     string
	RateLimiterType string
	Limits          LimitConfigMap
}

// Load reads the configuration file at the specified filepath and returns a NotificationService.
// It parses the JSON content of the file and populates the service's configurations.
func Load(filepath string) (*NotificationService, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var conf NotificationConfig
	if err := json.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	limits := make(map[string]*LimitConfig, len(conf.RateLimit.Limits))
	for _, config := range conf.RateLimit.Limits {
		limits[config.Type] = config
	}

	// Create a new NotificationService with the parsed configurations.
	service := &NotificationService{
		GatewayType:     conf.Gateway,
		RateLimiterType: conf.RateLimit.Type,
		Limits:          limits,
	}

	return service, nil
}

// WindowsSizeDuration returns the window size duration for the limit configuration.
func (conf *LimitConfig) WindowsSizeDuration() time.Duration {
	return time.Millisecond * time.Duration(conf.WSizeMs)
}
