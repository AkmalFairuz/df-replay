package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlaceBlock struct {
	Position  protocol.BlockPos
	BlockHash uint32
}

func (a *PlaceBlock) ID() uint8 {
	return IDPlaceBlock
}

func (a *PlaceBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	io.Uint32(&a.BlockHash)
}

func (a *PlaceBlock) Play(ctx *PlayContext) {
	pos := blockPosToCubePos(a.Position)
	b := internal.HashToBlock(a.BlockHash)
	prevBlock := ctx.Playback().Block(ctx.Tx(), pos)
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().SetBlock(ctx.Tx(), pos, prevBlock)
	})
	ctx.Playback().SetBlock(ctx.Tx(), pos, b)
	ctx.Playback().PlaySound(ctx.Tx(), pos.Vec3Centre(), sound.BlockPlace{Block: b})
}
