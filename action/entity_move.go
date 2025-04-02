package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type EntityMove struct {
	EntityID   uint32
	Position   mgl32.Vec3
	Yaw, Pitch float32
}

func (a *EntityMove) ID() uint8 {
	return IDEntityMove
}

func (a *EntityMove) Marshal(io protocol.IO) {
	io.Varuint32(&a.EntityID)
	io.Vec3(&a.Position)
	io.Float32(&a.Yaw)
	io.Float32(&a.Pitch)
}

func (a *EntityMove) Play(ctx *PlayContext) {
	prevPos, ok := ctx.Playback().EntityPosition(ctx.Tx(), a.EntityID)
	prevRot, _ := ctx.Playback().EntityRotation(ctx.Tx(), a.EntityID)
	if ok {
		ctx.OnReverse(func(ctx *PlayContext) {
			ctx.Playback().MoveEntity(ctx.Tx(), a.EntityID, prevPos, prevRot)
		})
	}
	ctx.Playback().MoveEntity(ctx.Tx(), a.EntityID, vec32To64(a.Position), cube.Rotation{float64(a.Yaw), float64(a.Pitch)})
}
