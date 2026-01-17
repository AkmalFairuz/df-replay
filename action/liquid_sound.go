package action

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	LiquidSoundTypeFill uint8 = iota
	LiquidSoundTypeEmpty
)

type LiquidSound struct {
	Liquid   Block
	Type     uint8
	Position protocol.BlockPos
}

func (a *LiquidSound) ID() uint8 {
	return IDLiquidSound
}

func (a *LiquidSound) Marshal(io protocol.IO) {
	protocol.Single(io, &a.Liquid)
	io.Uint8(&a.Type)
	io.BlockPos(&a.Position)
}

func (a *LiquidSound) Play(ctx *PlayContext) {
	liq, ok := a.Liquid.ToBlock().(world.Liquid)
	if !ok {
		return
	}
	pos := blockPosToCubePos(a.Position).Vec3Centre()
	do := func(ctx *PlayContext) {
		var s world.Sound
		switch a.Type {
		case LiquidSoundTypeFill:
			s = sound.BucketFill{Liquid: liq}
		case LiquidSoundTypeEmpty:
			s = sound.BucketEmpty{Liquid: liq}
		}
		ctx.Tx().PlaySound(pos, s)
	}
	ctx.OnReverse(do)
	do(ctx)
}
