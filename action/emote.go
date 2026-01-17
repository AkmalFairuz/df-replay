package action

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Emote struct {
	PlayerID uint32
	EmoteID  uuid.UUID
}

func (a *Emote) ID() uint8 {
	return IDEmote
}

func (a *Emote) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	io.UUID(&a.EmoteID)
}

func (a *Emote) Play(ctx *PlayContext) {
	do := func(ctx *PlayContext) {
		ctx.Playback().Emote(ctx.Tx(), a.PlayerID, a.EmoteID)
	}
	ctx.OnReverse(do)
	do(ctx)
}
