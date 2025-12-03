package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-project/serverless/components/operator/internal/file"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	notificationDelay = 1 * time.Second
)

type Config struct {
	ChartPath     string `envconfig:"default=/module-chart"`
	LogLevel      string `envconfig:"default=info" yaml:"logLevel"`
	LogFormat     string `envconfig:"default=json" yaml:"logFormat"`
	LogConfigPath string `envconfig:"optional"`
}

func GetConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

// LoadLogConfig - return cfg struct based on given path
func LoadLogConfig(path string) (Config, error) {
	cfg := Config{}

	cleanPath := filepath.Clean(path)
	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	return cfg, err
}

type CallbackFn func(Config)

// RunOnConfigChange - run callback functions when config is changed
func RunOnConfigChange(ctx context.Context, log *zap.SugaredLogger, path string, callbacks ...CallbackFn) {
	log.Info("config notifier started")

	for {
		// wait 1 sec not to burn out the container for example when any method below always ends with an error
		time.Sleep(notificationDelay)

		err := fireCallbacksOnConfigChange(ctx, log, path, callbacks...)
		if err != nil && errors.Is(err, context.Canceled) {
			log.Info("context canceled")
			return
		}
		if err != nil {
			log.Error(err)
		}
	}
}

func fireCallbacksOnConfigChange(ctx context.Context, log *zap.SugaredLogger, path string, callbacks ...CallbackFn) error {
	err := file.NotifyModification(ctx, path)
	if err != nil {
		return err
	}

	log.Info("config file change detected")

	cfg, err := LoadLogConfig(path)
	if err != nil {
		return err
	}

	log.Debugf("firing '%d' callbacks", len(callbacks))

	fireCallbacks(cfg, callbacks...)
	return nil
}

func fireCallbacks(cfg Config, funcs ...CallbackFn) {
	for i := range funcs {
		fn := funcs[i]
		fn(cfg)
	}
}
