package configs

import "time"

type (
	Unit         string
	LimitConfigs map[string]*Config
)

func (lc LimitConfigs) Get(typ string) *Config {
	return lc[typ]
}

var (
	Second Unit = "second"
	Minute Unit = "minute"
	Hour   Unit = "hour"
	Day    Unit = "day"
)

type Config struct {
	Limit int64 `json:"limit"`
	Unit  Unit  `json:"unit"`
}

func (conf *Config) UnitToDuration() time.Duration {
	switch conf.Unit {
	case Second:
		return time.Second
	case Minute:
		return time.Minute
	case Hour:
		return time.Hour
	case Day:
		return time.Hour * 24
	default:
		return time.Second
	}
}
