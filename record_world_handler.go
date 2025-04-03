package replay

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type RecordWorldHandler struct {
	world.NopHandler

	r *Recorder
}

func NewRecordWorldHandler(r *Recorder) *RecordWorldHandler {
	return &RecordWorldHandler{
		r: r,
	}
}

func (h *RecordWorldHandler) HandleEntitySpawn(tx *world.Tx, e world.Entity) {
	if _, ok := e.(*player.Player); ok {
		return
	}
	h.r.AddEntity(e)
}

func (h *RecordWorldHandler) HandleEntityDespawn(tx *world.Tx, e world.Entity) {
	if p, ok := e.(*player.Player); ok {
		h.r.RemovePlayer(p)
		return
	}
	h.r.RemoveEntity(e)
}
