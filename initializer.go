package replay

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server/entity"
	"sync/atomic"
)

var (
	initialized atomic.Bool
)

// Init should be called after items, blocks, and entities are registered. We don't
// use go init() because we want to control the order of initialization.
func Init() {
	if !initialized.CompareAndSwap(false, true) {
		panic("already initialized")
	}
	internal.ConstructBlockHashMappings()
	internal.ConstructItemHashMappings()

	entity.DefaultRegistry = entity.DefaultRegistry.Config().New(append(entity.DefaultRegistry.Types(), playerType, entityType))
	intercept.Hook(packetHandler{})
}
