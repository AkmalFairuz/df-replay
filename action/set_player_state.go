package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SetPlayerStateTypeVisibility uint8 = iota
	SetPlayerStateTypeSneaking
	SetPlayerStateTypeSprinting
	SetPlayerStateTypeGliding
	SetPlayerStateTypeUsingItem
	SetPlayerStateTypeSwimming
	SetPlayerStateTypeCrawling
)

type SetPlayerState struct {
	PlayerID uint32
	Type     uint8
	Value    bool
}

func (a *SetPlayerState) ID() uint8 {
	return IDSetPlayerState
}

func (a *SetPlayerState) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	io.Uint8(&a.Type)
	io.Bool(&a.Value)
}

func (a *SetPlayerState) Play(ctx *PlayContext) {
	switch a.Type {
	case SetPlayerStateTypeVisibility:
		prev := ctx.Playback().PlayerVisible(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerVisibility(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerVisibility(ctx.Tx(), a.PlayerID, a.Value)
	case SetPlayerStateTypeSneaking:
		prev := ctx.Playback().PlayerSneaking(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerSneaking(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerSneaking(ctx.Tx(), a.PlayerID, a.Value)
	case SetPlayerStateTypeUsingItem:
		prev := ctx.Playback().PlayerUsingItem(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerUsingItem(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerUsingItem(ctx.Tx(), a.PlayerID, a.Value)
	case SetPlayerStateTypeSprinting:
		prev := ctx.Playback().PlayerSprinting(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerSprinting(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerSprinting(ctx.Tx(), a.PlayerID, a.Value)
	case SetPlayerStateTypeGliding:
		prev := ctx.Playback().PlayerGliding(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerGliding(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerGliding(ctx.Tx(), a.PlayerID, a.Value)
	case SetPlayerStateTypeSwimming:
		prev := ctx.Playback().PlayerSwimming(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerSwimming(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerSwimming(ctx.Tx(), a.PlayerID, a.Value)
	case SetPlayerStateTypeCrawling:
		prev := ctx.Playback().PlayerCrawling(ctx.Tx(), a.PlayerID)
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().SetPlayerCrawling(ctx.Tx(), a.PlayerID, prev)
		})
		ctx.Playback().SetPlayerCrawling(ctx.Tx(), a.PlayerID, a.Value)
	}
}
