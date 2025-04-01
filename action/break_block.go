package action

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type BreakBlock struct {
	Position protocol.BlockPos
}

func (a *BreakBlock) ID() uint8 {
	return IDBreakBlock
}

func (a *BreakBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
}

func (a *BreakBlock) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	prevBlock := ctx.Playback().Block(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetBlock(ctx.Tx(), pos, prevBlock, 0)
	})
	ctx.Playback().SetBlock(ctx.Tx(), pos, block.Air{}, 0)
	ctx.Playback().AddParticle(ctx.Tx(), pos.Vec3Centre(), particle.BlockBreak{Block: prevBlock})
}
