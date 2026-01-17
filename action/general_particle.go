package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type GeneralParticle struct {
	Position   mgl32.Vec3
	ParticleID uint32
}

func (a *GeneralParticle) ID() uint8 {
	return IDGeneralParticle
}

func (a *GeneralParticle) Marshal(io protocol.IO) {
	io.Vec3(&a.Position)
	io.Varuint32(&a.ParticleID)
}

func (a *GeneralParticle) Play(ctx *PlayContext) {
	do := func(ctx *PlayContext) {
		if p, ok := internal.FromParticleID(a.ParticleID); ok {
			ctx.Playback().AddParticle(ctx.Tx(), vec32To64(a.Position), p)
		}
	}
	ctx.OnReverse(do)
	do(ctx)
}
