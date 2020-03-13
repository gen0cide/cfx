package cfx

import (
	"go.uber.org/fx"
)

// Module is the Fx provider that gives access to cfgfx.EnvContext and cfgfx.Container types.
var Module = fx.Provide(
	NewEnvContext,
	NewConfig,
)
