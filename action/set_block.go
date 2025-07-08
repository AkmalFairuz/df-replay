package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type SetBlock struct {
	Position protocol.BlockPos
	Block    Block
}

func (*SetBlock) ID() uint8 {
	return IDSetBlock
}

func (a *SetBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	protocol.Single(io, &a.Block)
}

func (a *SetBlock) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	prevBlock := ctx.Playback().Block(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetBlock(ctx.Tx(), pos, prevBlock)
	})
	ctx.Playback().SetBlock(ctx.Tx(), pos, a.Block.ToBlock())
}
