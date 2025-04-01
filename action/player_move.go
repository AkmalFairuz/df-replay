package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerMove struct {
	PlayerID uint32
	Position mgl32.Vec3
	Yaw      float32
	Pitch    float32
}

func (*PlayerMove) ID() uint8 {
	return IDPlayerMove
}

func (a *PlayerMove) Marshal(io protocol.IO) {
	io.Varuint32(&a.PlayerID)
	io.Vec3(&a.Position)
	io.Float32(&a.Yaw)
	io.Float32(&a.Pitch)
}

func (a *PlayerMove) Play(ctx *PlayContext) {
	prevPos, ok := ctx.Playback().PlayerPosition(ctx.Tx(), a.PlayerID)
	prevRot, ok2 := ctx.Playback().PlayerRotation(ctx.Tx(), a.PlayerID)
	if ok && ok2 {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().MovePlayer(ctx.Tx(), a.PlayerID, prevPos, prevRot)
		})
	}
	ctx.Playback().MovePlayer(ctx.Tx(), a.PlayerID, vec32To64(a.Position), cube.Rotation{float64(a.Yaw), float64(a.Pitch)})
}
