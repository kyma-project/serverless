package logging

import (
	"context"

	"github.com/kyma-project/manager-toolkit/logging/config"
	"go.uber.org/zap"
)

// ReconfigureOnConfigChange monitors config changes and updates log level dynamically.
// When log format changes, it triggers a graceful pod restart via the manager-toolkit implementation.
func ReconfigureOnConfigChange(ctx context.Context, log *zap.SugaredLogger, atomic zap.AtomicLevel, cfgPath string) {
	config.ReconfigureOnConfigChangeWithRestart(ctx, log, atomic, cfgPath)
}
