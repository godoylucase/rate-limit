package configs

import (
	"encoding/json"
	"os"
	"time"
)

type (
	Unit           string
	LimitConfigMap map[string]*LimitConfig
)

func (lcm LimitConfigMap) Get(typ string) *LimitConfig {
	return lcm[typ]
}

type NotificationConfig struct {
	Gateway   string           `json:"gateway"`
	RateLimit *RateLimitConfig `json:"rate_limit"`
}

type LimitConfig struct {
	Type    string `json:"type"`
	Limit   int64  `json:"limit"`
	WSizeMs int64  `json:"window_size_ms"`
}

type RateLimitConfig struct {
	Type   string         `json:"type"`
	Limits []*LimitConfig `json:"limits"`
}

type NotificationService struct {
	GatewayType     string
	RateLimiterType string
	Limits          LimitConfigMap
}

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

	return &NotificationService{
		GatewayType:     conf.Gateway,
		RateLimiterType: conf.RateLimit.Type,
		Limits:          limits,
	}, nil
}

func (conf *LimitConfig) WindowsSizeDuration() time.Duration {
	return time.Millisecond * time.Duration(conf.WSizeMs)
}
