package action

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type EntitySpawn struct {
	EntityID         uint32
	EntityIdentifier string
	NameTag          string
	Position         mgl32.Vec3
	Yaw, Pitch       uint16
	ExtraData        map[string]any
}

func (a *EntitySpawn) ID() uint8 {
	return IDEntitySpawn
}

func (a *EntitySpawn) Marshal(io protocol.IO) {
	io.Varuint32(&a.EntityID)
	io.String(&a.EntityIdentifier)
	io.String(&a.NameTag)
	io.Vec3(&a.Position)
	io.Uint16(&a.Yaw)
	io.Uint16(&a.Pitch)
	io.NBT(&a.ExtraData, nbt.LittleEndian)
}

func (a *EntitySpawn) Play(ctx *PlayContext) {
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().DespawnEntity(ctx.Tx(), a.EntityID)
	})
	ctx.Playback().SpawnEntity(ctx.Tx(), a.EntityID, a.EntityIdentifier, a.NameTag, vec32To64(a.Position), DecodeRotation16(a.Yaw, a.Pitch), a.ExtraData)
}
