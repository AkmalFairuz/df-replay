package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerAnimateSwing = iota
	PlayerAnimateSneak
	PlayerAnimateStopSneak
	PlayerAnimateHurt
	PlayerAnimateEating
	PlayerAnimateStartUsingItem
	PlayerAnimateStopUsingItem
	PlayerAnimateTotemUse
)

type PlayerAnimate struct {
	PlayerID  uint32
	Animation uint8
}

func (*PlayerAnimate) ID() uint8 {
	return IDPlayerAnimate
}

func (a *PlayerAnimate) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	io.Uint8(&a.Animation)
}

func (a *PlayerAnimate) Play(ctx *PlayContext) {
	switch a.Animation {
	case PlayerAnimateSwing:
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().DoPlayerSwingArm(ctx.Tx(), a.PlayerID)
		})
		ctx.Playback().DoPlayerSwingArm(ctx.Tx(), a.PlayerID)
	case PlayerAnimateSneak, PlayerAnimateStopSneak:
		prev := ctx.Playback().PlayerSneaking(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerSneaking(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerSneaking(ctx.Tx(), a.PlayerID, a.Animation == PlayerAnimateSneak)
	case PlayerAnimateHurt:
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().DoPlayerHurt(ctx.Tx(), a.PlayerID)
		})
		ctx.Playback().DoPlayerHurt(ctx.Tx(), a.PlayerID)
	case PlayerAnimateEating:
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().DoPlayerEating(ctx.Tx(), a.PlayerID)
		})
		ctx.Playback().DoPlayerEating(ctx.Tx(), a.PlayerID)
	case PlayerAnimateStartUsingItem, PlayerAnimateStopUsingItem:
		prev := ctx.Playback().PlayerUsingItem(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerUsingItem(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerUsingItem(ctx.Tx(), a.PlayerID, a.Animation == PlayerAnimateStartUsingItem)
	case PlayerAnimateTotemUse:
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().DoPlayerTotemUse(ctx.Tx(), a.PlayerID)
		})
		ctx.Playback().DoPlayerTotemUse(ctx.Tx(), a.PlayerID)
	default:
		// Unknown animation.
	}
}
