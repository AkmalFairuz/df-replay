package replay

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/entity"
)

// Init should be called after items, blocks, and entities are registered. We don't
// use go init() because we want to control the order of initialization.
func Init() {
	internal.ConstructBlockHashMappings()
	internal.ConstructItemHashMappings()

	entity.DefaultRegistry = entity.DefaultRegistry.Config().New(append(entity.DefaultRegistry.Types(), playerType))
}
