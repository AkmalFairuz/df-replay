package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type SetBlock struct {
	Position  protocol.BlockPos
	BlockHash uint64
	Layer     uint8
}

func (*SetBlock) ID() uint8 {
	return IDSetBlock
}

func (a *SetBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	io.Varuint64(&a.BlockHash)
	io.Uint8(&a.Layer)
}

func (a *SetBlock) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	prevBlock := ctx.Playback().Block(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetBlock(ctx.Tx(), pos, prevBlock, a.Layer)
	})
	ctx.Playback().SetBlock(ctx.Tx(), pos, internal.HashToBlock(a.BlockHash), a.Layer)
}
