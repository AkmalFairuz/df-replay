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

type PlayerHandler struct {
	player.NopHandler

	r *Recorder
}

func NewPlayerHandler(r *Recorder) PlayerHandler {
	return PlayerHandler{r: r}
}

func (h PlayerHandler) HandleMove(ctx *player.Context, pos mgl64.Vec3, rot cube.Rotation) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerMovement(ctx.Val(), pos, rot)
}

func (h PlayerHandler) HandleToggleSneak(ctx *player.Context, sneaking bool) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerSneaking(ctx.Val(), sneaking)
}

func (h PlayerHandler) HandleHeldSlotChange(ctx *player.Context, _, to int) {
	if ctx.Cancelled() {
		return
	}
	_, offHand := ctx.Val().HeldItems()
	mainHand, _ := ctx.Val().Inventory().Item(to)
	h.r.PushPlayerHandChange(ctx.Val(), mainHand, offHand)
	h.r.PushPlayerUsingItem(ctx.Val(), false)
}

func (h PlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlaceBlock(pos, b)
}

func (h PlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, _ *[]item.Stack, _ *int) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushBreakBlock(pos)
}

func (h PlayerHandler) HandleItemConsume(ctx *player.Context, _ item.Stack) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerUsingItem(ctx.Val(), false)
}

func (h PlayerHandler) HandleItemRelease(ctx *player.Context, _ item.Stack, _ time.Duration) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerUsingItem(ctx.Val(), false)
}

func (h PlayerHandler) HandleItemUse(ctx *player.Context) {
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

func (h PlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSkinChange(ctx.Val(), *skin)
}

func (h PlayerHandler) HandleHurt(ctx *player.Context, _ *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerHurt(ctx.Val())
}

func (h PlayerHandler) HandlePunchAir(ctx *player.Context) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerSwingArm(ctx.Val())
}

func (h PlayerHandler) HandleItemUseOnEntity(ctx *player.Context, _ world.Entity) {
	if ctx.Cancelled() {
		return
	}
	// TODO: this handler may not call SwingArm!
	h.r.PushPlayerSwingArm(ctx.Val())
}

func (h PlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, _ cube.Face, _ mgl64.Vec3) {
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
