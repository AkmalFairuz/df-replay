package action

import (
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlaceBlock struct {
	Position protocol.BlockPos
	Block    Block
}

func (a *PlaceBlock) ID() uint8 {
	return IDPlaceBlock
}

func (a *PlaceBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	protocol.Single(io, &a.Block)
}

func (a *PlaceBlock) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	b := a.Block.ToBlock()
	prevBlock := ctx.Playback().Block(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetBlock(ctx.Tx(), pos, prevBlock)
	})
	ctx.Playback().SetBlock(ctx.Tx(), pos, b)
	ctx.Playback().PlaySound(ctx.Tx(), pos.Vec3Centre(), sound.BlockPlace{Block: b})
}
