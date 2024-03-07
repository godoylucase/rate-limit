// Package configs provides functionality for loading and parsing notification configurations.
package configs

import (
	"encoding/json"
	"fmt"
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

// JsonConfiguration represents the configuration for the notification service.
type JsonConfiguration struct {
	Redis     *RedisConfig     `json:"redis"`
	RateLimit *RateLimitConfig `json:"rate_limit"`
}

type RedisConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (rc *RedisConfig) Address() string {
	if rc.Port != 0 {
		return fmt.Sprintf("%v:%v", rc.Host, rc.Port)
	}
	return rc.Host
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
	RedisAddr       string
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

	var jsonConf JsonConfiguration
	if err := json.Unmarshal(content, &jsonConf); err != nil {
		return nil, err
	}

	limits := make(map[string]*LimitConfig, len(jsonConf.RateLimit.Limits))
	for _, config := range jsonConf.RateLimit.Limits {
		limits[config.Type] = config
	}

	// Create a new NotificationService with the parsed configurations.
	service := &NotificationService{
		RedisAddr:       jsonConf.Redis.Address(),
		RateLimiterType: jsonConf.RateLimit.Type,
		Limits:          limits,
	}

	return service, nil
}

// WindowsSizeDuration returns the window size duration for the limit configuration.
func (conf *LimitConfig) WindowsSizeDuration() time.Duration {
	return time.Millisecond * time.Duration(conf.WSizeMs)
}
