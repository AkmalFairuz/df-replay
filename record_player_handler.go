package replay

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

type RecordPlayerHandler struct {
	player.NopHandler

	r *Recorder
}

func NewRecordPlayerHandler(r *Recorder) *RecordPlayerHandler {
	return &RecordPlayerHandler{r: r}
}

func (h *RecordPlayerHandler) HandleMove(ctx *player.Context, pos mgl64.Vec3, rot cube.Rotation) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerMovement(ctx.Val(), pos, rot)
}

func (h *RecordPlayerHandler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	if ctx.Cancelled() {
		return
	}
	// TODO: don't use movement
	h.r.PushPlayerMovement(ctx.Val(), pos, ctx.Val().Rotation())
}

func (h *RecordPlayerHandler) HandleToggleSneak(ctx *player.Context, sneaking bool) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerSneaking(ctx.Val(), sneaking)
}

func (h *RecordPlayerHandler) HandleHeldSlotChange(ctx *player.Context, _, to int) {
	if ctx.Cancelled() {
		return
	}
	_, offHand := ctx.Val().HeldItems()
	mainHand, _ := ctx.Val().Inventory().Item(to)
	h.r.PushPlayerHandChange(ctx.Val(), mainHand, offHand)
	h.r.PushPlayerUsingItem(ctx.Val(), false)
}

func (h *RecordPlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlaceBlock(pos, b)
	h.r.PushPlayerSwingArm(ctx.Val())
}

func (h *RecordPlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, _ *[]item.Stack, _ *int) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushBreakBlock(pos)
	h.r.PushPlayerSwingArm(ctx.Val())
}

func (h *RecordPlayerHandler) HandleItemConsume(ctx *player.Context, _ item.Stack) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerUsingItem(ctx.Val(), false)
}

func (h *RecordPlayerHandler) HandleItemRelease(ctx *player.Context, _ item.Stack, _ time.Duration) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerUsingItem(ctx.Val(), false)
}

func (h *RecordPlayerHandler) HandleItemUse(ctx *player.Context) {
	if ctx.Cancelled() {
		return
	}
	mainHand, _ := ctx.Val().HeldItems()
	switch mainHand.Item().(type) {
	case item.Releasable:
		if !player_canRelease(ctx.Val()) {
			return
		}
		h.r.PushPlayerUsingItem(ctx.Val(), true)
	case item.Consumable:
		h.r.PushPlayerEating(ctx.Val())
	default:
		// handle other usable item, like crossbow
	}
}

func (h *RecordPlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSkinChange(ctx.Val(), *skin)
}

func (h *RecordPlayerHandler) HandleHurt(ctx *player.Context, _ *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerHurt(ctx.Val())
}

func (h *RecordPlayerHandler) HandlePunchAir(ctx *player.Context) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerSwingArm(ctx.Val())
}

func (h *RecordPlayerHandler) HandleItemUseOnEntity(ctx *player.Context, _ world.Entity) {
	if ctx.Cancelled() {
		return
	}
	// TODO: this handler may not call SwingArm!
	h.r.PushPlayerSwingArm(ctx.Val())
}

func (h *RecordPlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, _ cube.Face, _ mgl64.Vec3) {
	if ctx.Cancelled() {
		return
	}
	b := ctx.Val().Tx().Block(pos)
	if _, ok := b.(block.Activatable); ok {
		h.r.PushPlayerSwingArm(ctx.Val())
		return
	}

	mainHand, _ := ctx.Val().HeldItems()
	if _, ok := mainHand.Item().(item.UsableOnBlock); ok {
		h.r.PushPlayerSwingArm(ctx.Val())
		return
	}
}
