package action

import (
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerSkin struct {
	PlayerID        uint32
	SkinWidth       uint32
	SkinHeight      uint32
	SkinData        []byte
	HasCape         bool
	CapeWidth       uint32
	CapeHeight      uint32
	CapeData        []byte
	GeometryName    string
	HasGeometryData bool
	GeometryData    []byte
}

func (a *PlayerSkin) ID() uint8 {
	return IDPlayerSkin
}

func (a *PlayerSkin) Marshal(io protocol.IO) {
	io.Uint32(&a.PlayerID)
	io.Uint32(&a.SkinWidth)
	io.Uint32(&a.SkinHeight)
	io.Bytes(&a.SkinData)
	io.Bool(&a.HasCape)
	if a.HasCape {
		io.Uint32(&a.CapeWidth)
		io.Uint32(&a.CapeHeight)
		io.Bytes(&a.CapeData)
	}
	io.String(&a.GeometryName)
	io.Bool(&a.HasGeometryData)
	if a.HasGeometryData {
		io.Bytes(&a.GeometryData)
	}
}

func (a *PlayerSkin) Play(ctx *PlayContext) {
	prevSkin, ok := ctx.Playback().PlayerSkin(a.PlayerID)
	if ok {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().UpdatePlayerSkin(ctx.Tx(), a.PlayerID, prevSkin)
		})
	}
	sk := skin.New(int(a.SkinWidth), int(a.SkinHeight))
	sk.Pix = a.SkinData
	if a.HasCape {
		sk.Cape = skin.NewCape(int(a.CapeWidth), int(a.CapeHeight))
		sk.Cape.Pix = a.CapeData
	}
	sk.ModelConfig.Default = a.GeometryName
	if a.HasGeometryData {
		sk.Model = a.GeometryData
	}
	ctx.Playback().UpdatePlayerSkin(ctx.Tx(), a.PlayerID, sk)
}
