package action

import (
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type SetPlayerVisibleEffects struct {
	PlayerID uint32
	Effects  []uint8
}

func (a *SetPlayerVisibleEffects) ID() uint8 {
	return IDSetPlayerVisibleEffects
}

func (a *SetPlayerVisibleEffects) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	protocol.FuncSlice(io, &a.Effects, io.Uint8)
}

func (a *SetPlayerVisibleEffects) Play(ctx *PlayContext) {
	prev, _ := ctx.Playback().PlayerVisibleEffects(ctx.Tx(), a.PlayerID)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetPlayerVisibleEffects(ctx.Tx(), a.PlayerID, prev)
	})
	ctx.Playback().SetPlayerVisibleEffects(ctx.Tx(), a.PlayerID, lo.Map(a.Effects, func(e uint8, _ int) int {
		return int(e)
	}))
}
