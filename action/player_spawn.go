package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerSpawn struct {
	PlayerID   uint32
	PlayerName string

	Position   mgl32.Vec3
	Yaw, Pitch float32

	Helmet, Chestplate, Leggings, Boots Item
	MainHand, OffHand                   Item
}

func (*PlayerSpawn) ID() uint8 {
	return IDPlayerSpawn
}

func (a *PlayerSpawn) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	io.String(&a.PlayerName)
	io.Vec3(&a.Position)
	io.Float32(&a.Yaw)
	io.Float32(&a.Pitch)
	protocol.Single(io, &a.Helmet)
	protocol.Single(io, &a.Chestplate)
	protocol.Single(io, &a.Leggings)
	protocol.Single(io, &a.Boots)
	protocol.Single(io, &a.MainHand)
	protocol.Single(io, &a.OffHand)
}

func (a *PlayerSpawn) Play(ctx *PlayContext) {
	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().DespawnPlayer(ctx.Tx(), a.PlayerID)
	})
	ctx.Playback().SpawnPlayer(
		ctx.Tx(), a.PlayerName, a.PlayerID, vec32To64(a.Position),
		cube.Rotation{float64(a.Yaw), float64(a.Pitch)},
		[4]item.Stack{a.Helmet.ToStack(), a.Chestplate.ToStack(), a.Leggings.ToStack(), a.Boots.ToStack()},
		[2]item.Stack{a.MainHand.ToStack(), a.OffHand.ToStack()})
}
