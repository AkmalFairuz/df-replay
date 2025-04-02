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
	data.Data = &replayEntityData{
		Identifier: m["Identifier"].(string),
		ExtraData:  m["ExtraData"].(map[string]any),
	}
}

func (et etype) EncodeNBT(data *world.EntityData) map[string]any {
	return data.Data.(map[string]any)
}

type replayEntityData struct {
	Identifier string
	ExtraData  map[string]any
}
