package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerNameTagUpdate struct {
	PlayerID uint32
	NameTag  string
}

func (a *PlayerNameTagUpdate) ID() uint8 {
	return IDPlayerNameTagUpdate
}

func (a *PlayerNameTagUpdate) Marshal(io protocol.IO) {
	io.Uint32(&a.PlayerID)
	io.String(&a.NameTag)
}

func (a *PlayerNameTagUpdate) Play(ctx *PlayContext) {
	prev := ctx.Playback().PlayerNameTag(ctx.Tx(), a.PlayerID)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetPlayerNameTag(ctx.Tx(), a.PlayerID, prev)
	})
	ctx.Playback().SetPlayerNameTag(ctx.Tx(), a.PlayerID, a.NameTag)
}
