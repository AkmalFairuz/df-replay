package replay

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
)

type Entity struct {
	id         uint32
	identifier string
	extraData  map[string]any
	h          *world.EntityHandle
}

var entityType = etype{}

type etype struct{}

func (et etype) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return entity.Open(tx, handle, data)
}

func (et etype) EncodeEntity() string {
	return "replay_entity"
}

func (et etype) BBox(_ world.Entity) cube.BBox {
	// TODO: implement correct bbox
	return cube.Box(-0.25, 0, -0.25, 0.25, 0.5, 0.25)
}

func (et etype) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := &entityBehaviourConfig{
		Identifier: m["Identifier"].(string),
		ExtraData:  m["ExtraData"].(map[string]any),
	}
	data.Data = &entityBehaviour{
		identifier: conf.Identifier,
		extraData:  conf.ExtraData,
	}
}

func (et etype) EncodeNBT(data *world.EntityData) map[string]any {
	behaviour := data.Data.(*entityBehaviour)
	return map[string]any{
		"Identifier": behaviour.identifier,
		"ExtraData":  behaviour.extraData,
	}
}
