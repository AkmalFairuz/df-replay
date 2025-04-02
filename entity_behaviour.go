package replay

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
)

type entityBehaviourConfig struct {
	Identifier string
	ExtraData  map[string]any
}

func (conf entityBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

func (conf entityBehaviourConfig) New() *entityBehaviour {
	ret := &entityBehaviour{
		identifier: conf.Identifier,
		extraData:  conf.ExtraData,
	}
	return ret
}

type entityBehaviour struct {
	identifier string
	extraData  map[string]any
}

func (b *entityBehaviour) Tick(*entity.Ent, *world.Tx) *entity.Movement {
	return nil
}
