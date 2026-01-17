package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BlockParticleTypeBreak uint8 = iota
	BlockParticleTypePunching
)

type BlockParticle struct {
	Position protocol.BlockPos
	Block    Block
	Type     uint8
	Face     uint8
}

func (a *BlockParticle) ID() uint8 {
	return IDBlockParticle
}

func (a *BlockParticle) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	protocol.Single(io, &a.Block)
	io.Uint8(&a.Type)
	if a.Type == BlockParticleTypePunching {
		io.Uint8(&a.Face)
	}
}

func (a *BlockParticle) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position).Vec3Centre()
	b := a.Block.ToBlock()
	do := func(ctx *PlayContext) {
		switch a.Type {
		case BlockParticleTypeBreak:
			ctx.Tx().AddParticle(pos, particle.BlockBreak{Block: b})
		case BlockParticleTypePunching:
			ctx.Tx().AddParticle(pos, particle.PunchBlock{Block: b, Face: cube.Face(a.Face)})
		}
	}
	ctx.OnReverse(do)
	do(ctx)
}
