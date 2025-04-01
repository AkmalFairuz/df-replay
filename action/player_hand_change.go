package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerHandChange struct {
	PlayerID uint32
	MainHand Item
	OffHand  Item
}

func (a *PlayerHandChange) ID() uint8 {
	return IDPlayerHandChange
}

func (a *PlayerHandChange) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	protocol.Single(io, &a.MainHand)
	protocol.Single(io, &a.OffHand)
}

func (a *PlayerHandChange) Play(ctx *PlayContext) {
	mainHand, offHand, ok := ctx.Playback().PlayerHeldItems(ctx.Tx(), a.PlayerID)
	if ok {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().UpdatePlayerHeldItems(ctx.Tx(), a.PlayerID, mainHand, offHand)
		})
	}
	ctx.Playback().UpdatePlayerHeldItems(ctx.Tx(), a.PlayerID, a.MainHand.ToStack(), a.OffHand.ToStack())
}
