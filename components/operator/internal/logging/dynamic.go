package logging

import (
	"context"

	"github.com/kyma-project/manager-toolkit/logging/config"
	"github.com/kyma-project/manager-toolkit/logging/logger"
	"go.uber.org/zap"
)

// ReconfigureOnConfigChangeWithRestart monitors config changes and updates log level dynamically.
// When log format changes, it triggers a graceful pod restart by exiting the process.
// This implements the restart logic locally on the operator side, similar to buildless-serverless.
func ReconfigureOnConfigChangeWithRestart(ctx context.Context, log *zap.SugaredLogger, atomic zap.AtomicLevel, cfgPath string, onFormatChange func()) {
	log.Info("Starting log configuration watcher")
	defer log.Info("Log configuration watcher stopped")

	config.RunOnConfigChange(ctx, log, cfgPath, func(cfg config.Config) {
		log.Infof("Log configuration changed: level=%s, format=%s", cfg.LogLevel, cfg.LogFormat)

		// Validate and map log level using manager-toolkit's MapLevel
		parsedLevel, err := logger.MapLevel(cfg.LogLevel)
		if err != nil {
			log.Errorf("Failed to parse log level %s: %v", cfg.LogLevel, err)
			return
		}

		// Convert to zapcore.Level and update log level dynamically
		zapLevel, err := parsedLevel.ToZapLevel()
		if err != nil {
			log.Errorf("Failed to convert log level %s to zap level: %v", cfg.LogLevel, err)
			return
		}
		atomic.SetLevel(zapLevel)
		log.Infof("Log level updated to: %s", cfg.LogLevel)

		// Validate and map log format using manager-toolkit's MapFormat
		// This handles normalization (e.g., "text" -> "console")
		newFormat, err := logger.MapFormat(cfg.LogFormat)
		if err != nil {
			log.Errorf("Failed to parse log format %s: %v", cfg.LogFormat, err)
			return
		}

		// Check if log format has changed - if so, trigger restart
		// This is the local restart logic, not delegated to manager-toolkit
		currentFormat := config.GetCurrentFormat()
		currentFormatMapped, err := logger.MapFormat(currentFormat)
		if err != nil {
			log.Errorf("Failed to parse current log format %s: %v", currentFormat, err)
			return
		}

		if currentFormatMapped != newFormat {
			log.Infof("Log format changed from %s to %s, triggering graceful restart", currentFormat, cfg.LogFormat)
			// Trigger restart callback
			if onFormatChange != nil {
				onFormatChange()
			}
			return
		}

		log.Infof("logger reconfigured with level '%s' and format '%s'", cfg.LogLevel, cfg.LogFormat)
	})
}
