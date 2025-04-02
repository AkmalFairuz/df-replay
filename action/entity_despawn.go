package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type EntityDespawn struct {
	EntityID uint32
}

func (a *EntityDespawn) ID() uint8 {
	return IDEntityDespawn
}

func (a *EntityDespawn) Marshal(io protocol.IO) {
	io.Varuint32(&a.EntityID)
}

func (a *EntityDespawn) Play(ctx *PlayContext) {
	prevPos, ok := ctx.Playback().EntityPosition(ctx.Tx(), a.EntityID)
	prevRot, _ := ctx.Playback().EntityRotation(ctx.Tx(), a.EntityID)
	prevExtraData, _ := ctx.Playback().EntityExtraData(a.EntityID)
	prevEntityIdentifier, _ := ctx.Playback().EntityIdentifier(a.EntityID)
	if ok {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SpawnEntity(ctx.Tx(), a.EntityID, prevEntityIdentifier, prevPos, prevRot, prevExtraData)
		})
	}
	ctx.Playback().DespawnEntity(ctx.Tx(), a.EntityID)
}
