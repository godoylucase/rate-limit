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

type LimitConfig struct {
	Type    string `json:"type"`
	Limit   int64  `json:"limit"`
	WSizeMs int64  `json:"window_size_ms"`
}

type LimitConfigs struct {
	Configs []*LimitConfig `json:"configs"`
}

func LoadLimitConfigs(filepath string) (LimitConfigMap, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var limitConfigs LimitConfigs
	if err := json.Unmarshal(content, &limitConfigs); err != nil {
		return nil, err
	}

	values := make(map[string]*LimitConfig, len(limitConfigs.Configs))
	for _, config := range limitConfigs.Configs {
		values[config.Type] = config
	}

	return values, nil
}

func (conf *LimitConfig) WindowsSizeDuration() time.Duration {
	return time.Millisecond * time.Duration(conf.WSizeMs)
}
