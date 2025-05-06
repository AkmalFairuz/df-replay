package replay

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
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

func (h *RecordWorldHandler) HandleLiquidFlow(ctx *world.Context, from, into cube.Pos, liquid world.Liquid, replaced world.Block) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSetLiquid(into, liquid)
}

func (h *RecordWorldHandler) HandleLiquidDecay(ctx *world.Context, pos cube.Pos, before, after world.Liquid) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSetLiquid(pos, after)
}

func (h *RecordWorldHandler) HandleLiquidHarden(ctx *world.Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock world.Block) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSetBlock(hardenedPos, newBlock)
}

func (h *RecordWorldHandler) HandleLeavesDecay(ctx *world.Context, pos cube.Pos) {
	if ctx.Cancelled() {
		return
	}
	b := ctx.Val().Block(pos)
	b2, ok := b.(block.Leaves)
	if ok {
		b2.ShouldUpdate = false
		h.r.PushSetBlock(pos, b2)
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
	h.r.entityMovementRecorder.removeLastMovement(e)
	h.r.RemoveEntity(e)

	h.r.removeLastEntityMovement(e)
}

func (h *RecordWorldHandler) HandleExplosion(ctx *world.Context, _ mgl64.Vec3, _ *[]world.Entity, blocks *[]cube.Pos, _ *float64, _ *bool) {
	if ctx.Cancelled() {
		return
	}
	for _, b := range *blocks {
		h.r.PushSetBlock(b, block.Air{})
	}
}

func (h *RecordWorldHandler) HandleCropTrample(ctx *world.Context, pos cube.Pos) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSetBlock(pos, block.Dirt{})
}
