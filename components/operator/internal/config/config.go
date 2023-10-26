package config

import "github.com/vrischmann/envconfig"

type Config struct {
	ChartPath                  string `envconfig:"default=/module-chart"`
	ServerlessManagerNamespace string `envconfig:"default=default"`
}

func GetConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err

}
