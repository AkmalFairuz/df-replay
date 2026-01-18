package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EntityAnimateFireworkExplosion = iota
	EntityAnimateArrowShake
)

type EntityAnimate struct {
	EntityID  uint32
	Animation uint8
}

func (a *EntityAnimate) ID() uint8 {
	return IDEntityAnimate
}

func (a *EntityAnimate) Marshal(io protocol.IO) {
	io.Varuint32(&a.EntityID)
	io.Uint8(&a.Animation)
}

func (a *EntityAnimate) Play(ctx *PlayContext) {
	switch a.Animation {
	case EntityAnimateFireworkExplosion:
		do := func(ctx *PlayContext) {
			ctx.Playback().DoFireworkExplosion(ctx.Tx(), a.EntityID)
		}
		ctx.OnReverse(do)
		do(ctx)
	case EntityAnimateArrowShake:
		do := func(ctx *PlayContext) {
			ctx.Playback().DoArrowShake(ctx.Tx(), a.EntityID)
		}
		ctx.OnReverse(do)
		do(ctx)
	}
}
