package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type GeneralSound struct {
	Position mgl32.Vec3
	SoundID  uint32
}

func (a *GeneralSound) ID() uint8 {
	return IDGeneralSound
}

func (a *GeneralSound) Marshal(io protocol.IO) {
	io.Vec3(&a.Position)
	io.Varuint32(&a.SoundID)
}

func (a *GeneralSound) Play(ctx *PlayContext) {
	s, ok := internal.FromSoundID(a.SoundID)
	if !ok {
		return
	}
	pos := vec32To64(a.Position)
	do := func(ctx *PlayContext) {
		ctx.Tx().PlaySound(pos, s)
	}
	ctx.OnReverse(do)
	do(ctx)
}
