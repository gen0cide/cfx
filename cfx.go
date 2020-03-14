package cfx

import (
	"go.uber.org/fx"
)

// Module is the Fx provider that gives access to cfgfx.EnvContext and cfgfx.Container types.
// Note: You should use cfx.NewFXEnvContext("PREFIX") to populate a constructor for EnvContext types.
// If you wish to leave "PREFIX" empty, you can - the default prefix is "CFX".
var Module = fx.Provide(
	NewConfig,
)
