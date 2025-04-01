package replay

import "github.com/df-mc/dragonfly/server/world"

type Entity struct {
	id uint32
	h  *world.EntityHandle
}
