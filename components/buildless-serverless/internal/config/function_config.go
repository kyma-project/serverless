package config

import "time"

type FunctionConfig struct {
	ImageNodeJs20                string
	ImagePython312               string
	RequeueDuration              time.Duration `envconfig:"default=1m"`
	FunctionReadyRequeueDuration time.Duration `envconfig:"default=5m"`
}
