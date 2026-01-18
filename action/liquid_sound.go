package action

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	LiquidSoundTypeFill uint8 = iota
	LiquidSoundTypeEmpty
)

type LiquidSound struct {
	IsWater  bool
	Type     uint8
	Position protocol.BlockPos
}

func (a *LiquidSound) ID() uint8 {
	return IDLiquidSound
}

func (a *LiquidSound) Marshal(io protocol.IO) {
	io.Bool(&a.IsWater)
	io.Uint8(&a.Type)
	io.BlockPos(&a.Position)
}

func (a *LiquidSound) Play(ctx *PlayContext) {
	var liq world.Liquid = block.Water{}
	if !a.IsWater {
		liq = block.Lava{}
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
