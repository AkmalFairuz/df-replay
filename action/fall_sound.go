package action

import (
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type FallSound struct {
	Position protocol.BlockPos
	Distance float32
}

func (a *FallSound) ID() uint8 {
	return IDFallSound
}

func (a *FallSound) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	io.Float32(&a.Distance)
}

func (a *FallSound) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position).Vec3Centre()
	do := func(ctx *PlayContext) {
		ctx.Tx().PlaySound(pos, sound.Fall{Distance: float64(a.Distance)})
	}
	ctx.OnReverse(do)
	do(ctx)
}
