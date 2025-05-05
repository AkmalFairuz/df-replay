package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerArmorChange struct {
	PlayerID   uint32
	Helmet     Item
	Chestplate Item
	Leggings   Item
	Boots      Item
}

func (a *PlayerArmorChange) ID() uint8 {
	return IDPlayerArmorChange
}

func (a *PlayerArmorChange) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	protocol.Single(io, &a.Helmet)
	protocol.Single(io, &a.Chestplate)
	protocol.Single(io, &a.Leggings)
	protocol.Single(io, &a.Boots)
}

func (a *PlayerArmorChange) Play(ctx *PlayContext) {
	helmet, chestplate, leggings, boots, ok := ctx.Playback().PlayerArmours(ctx.Tx(), a.PlayerID)
	if ok {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().UpdatePlayerArmours(ctx.Tx(), a.PlayerID, helmet, chestplate, leggings, boots)
		})
	}
	ctx.Playback().UpdatePlayerArmours(ctx.Tx(), a.PlayerID, a.Helmet.ToStack(), a.Chestplate.ToStack(), a.Leggings.ToStack(), a.Boots.ToStack())
}
