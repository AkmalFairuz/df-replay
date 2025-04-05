package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type SetLiquid struct {
	Position   protocol.BlockPos
	LiquidHash uint32
}

func (a *SetLiquid) ID() uint8 {
	return IDSetLiquid
}

func (a *SetLiquid) Marshal(io protocol.IO) {
	io.Uint32(&a.LiquidHash)
	io.BlockPos(&a.Position)
}

func (a *SetLiquid) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	prev, _ := ctx.Playback().Liquid(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetLiquid(ctx.Tx(), pos, prev)
	})
	l := internal.HashToBlock(a.LiquidHash)
	if liq, ok := l.(world.Liquid); ok {
		ctx.Playback().SetLiquid(ctx.Tx(), pos, liq)
	} else if l == (block.Air{}) {
		ctx.Playback().SetLiquid(ctx.Tx(), pos, nil)
	}
}
