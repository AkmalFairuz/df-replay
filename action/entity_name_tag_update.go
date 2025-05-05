package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type EntityNameTagUpdate struct {
	EntityID uint32
	NameTag  string
}

func (a *EntityNameTagUpdate) ID() uint8 {
	return IDEntityNameTagUpdate
}

func (a *EntityNameTagUpdate) Marshal(io protocol.IO) {
	io.Varuint32(&a.EntityID)
	io.String(&a.NameTag)
}

func (a *EntityNameTagUpdate) Play(ctx *PlayContext) {
	prev := ctx.Playback().EntityNameTag(ctx.Tx(), a.EntityID)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetEntityNameTag(ctx.Tx(), a.EntityID, prev)
	})
	ctx.Playback().SetEntityNameTag(ctx.Tx(), a.EntityID, a.NameTag)
}
