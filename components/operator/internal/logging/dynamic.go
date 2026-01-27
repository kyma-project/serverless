package logging

import (
	"context"

	"github.com/kyma-project/manager-toolkit/logging/config"
	"github.com/kyma-project/manager-toolkit/logging/logger"
	"go.uber.org/zap"
)

// ReconfigureOnConfigChange monitors config changes and updates log level dynamically.
// This is a thin wrapper around the manager-toolkit implementation.
func ReconfigureOnConfigChange(ctx context.Context, log *zap.SugaredLogger, atomic zap.AtomicLevel, cfgPath string) {
	config.RunOnConfigChange(ctx, log, cfgPath, func(cfg config.Config) {
		// Update log level dynamically
		level, err := logger.MapLevel(cfg.LogLevel)
		if err != nil {
			log.Error(err)
			return
		}
		zapLevel, err := level.ToZapLevel()
		if err != nil {
			log.Error(err)
			return
		}
		atomic.SetLevel(zapLevel)

		log.Infof("logger reconfigured with level '%s'. format changes require pod restart (requested format: '%s')", cfg.LogLevel, cfg.LogFormat)
	})
}
