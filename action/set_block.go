package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type SetBlock struct {
	Position  protocol.BlockPos
	BlockHash uint32
}

func (*SetBlock) ID() uint8 {
	return IDSetBlock
}

func (a *SetBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	io.Uint32(&a.BlockHash)
}

func (a *SetBlock) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	prevBlock := ctx.Playback().Block(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetBlock(ctx.Tx(), pos, prevBlock)
	})
	ctx.Playback().SetBlock(ctx.Tx(), pos, internal.HashToBlock(a.BlockHash))
}
