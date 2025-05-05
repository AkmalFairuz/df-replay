package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type ChestUpdate struct {
	Position protocol.BlockPos
	Open     bool
}

func (*ChestUpdate) ID() uint8 {
	return IDChestUpdate
}

func (c *ChestUpdate) Marshal(io protocol.IO) {
	io.BlockPos(&c.Position)
	io.Bool(&c.Open)
}

func (c *ChestUpdate) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(c.Position)
	state := ctx.Playback().ChestState(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().UpdateChestState(ctx.Tx(), pos, state)
	})
	ctx.Playback().UpdateChestState(ctx.Tx(), pos, c.Open)
}
