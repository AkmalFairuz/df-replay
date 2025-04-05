package action

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerDespawn struct {
	PlayerID uint32
}

func (*PlayerDespawn) ID() uint8 {
	return IDPlayerDespawn
}

func (a *PlayerDespawn) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
}

func (a *PlayerDespawn) Play(ctx *PlayContext) {
	prevHelmet, prevChestplate, prevLeggings, prevBoots, _ := ctx.Playback().PlayerArmours(ctx.Tx(), a.PlayerID)
	prevMainHand, prevOffHand, _ := ctx.Playback().PlayerHeldItems(ctx.Tx(), a.PlayerID)
	prevName := ctx.Playback().PlayerName(a.PlayerID)
	prevRot, _ := ctx.Playback().PlayerRotation(ctx.Tx(), a.PlayerID)
	prevPos, ok := ctx.Playback().PlayerPosition(ctx.Tx(), a.PlayerID)
	prevNameTag := ctx.Playback().PlayerNameTag(ctx.Tx(), a.PlayerID)
	if ok {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SpawnPlayer(
				ctx.Tx(), prevName, prevNameTag, a.PlayerID, prevPos, prevRot,
				[4]item.Stack{prevHelmet, prevChestplate, prevLeggings, prevBoots},
				[2]item.Stack{prevMainHand, prevOffHand})
		})
	}
	ctx.Playback().DespawnPlayer(ctx.Tx(), a.PlayerID)
}
