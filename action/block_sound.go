package action

import (
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BlockSoundTypeBreaking uint8 = iota
	BlockSoundTypePlace
)

type BlockSound struct {
	Position protocol.BlockPos
	Block    Block
	Type     uint8
}

func (a *BlockSound) ID() uint8 {
	return IDBlockSound
}

func (a *BlockSound) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	protocol.Single(io, &a.Block)
	io.Uint8(&a.Type)
}

func (a *BlockSound) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position).Vec3Centre()
	b := a.Block.ToBlock()
	do := func(ctx *PlayContext) {
		switch a.Type {
		case BlockSoundTypeBreaking:
			ctx.Playback().PlaySound(ctx.Tx(), pos, sound.BlockBreaking{Block: b})
		case BlockSoundTypePlace:
			ctx.Playback().PlaySound(ctx.Tx(), pos, sound.BlockPlace{Block: b})
		}
	}
	ctx.OnReverse(do)
	do(ctx)
}
