package action

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EntityDeltaMoveHasXFlag = 1 << iota
	EntityDeltaMoveHasYFlag
	EntityDeltaMoveHasZFlag
	EntityDeltaMoveHasYawFlag
	EntityDeltaMoveHasPitchFlag
)

type EntityDeltaMove struct {
	Flags      uint8
	EntityID   uint32
	Position   mgl32.Vec3
	Yaw, Pitch uint16
}

func (*EntityDeltaMove) ID() uint8 {
	return IDEntityDeltaMove
}

func (a *EntityDeltaMove) HasX() bool {
	return a.Flags&EntityDeltaMoveHasXFlag != 0
}

func (a *EntityDeltaMove) HasY() bool {
	return a.Flags&EntityDeltaMoveHasYFlag != 0
}

func (a *EntityDeltaMove) HasZ() bool {
	return a.Flags&EntityDeltaMoveHasZFlag != 0
}

func (a *EntityDeltaMove) HasYaw() bool {
	return a.Flags&EntityDeltaMoveHasYawFlag != 0
}

func (a *EntityDeltaMove) HasPitch() bool {
	return a.Flags&EntityDeltaMoveHasPitchFlag != 0
}

func (a *EntityDeltaMove) Marshal(io protocol.IO) {
	io.Uint8(&a.Flags)
	io.Varuint32(&a.EntityID)
	if a.HasX() {
		io.Float32(&a.Position[0])
	} else {
		a.Position[0] = 0
	}

	if a.HasY() {
		io.Float32(&a.Position[1])
	} else {
		a.Position[1] = 0
	}

	if a.HasZ() {
		io.Float32(&a.Position[2])
	} else {
		a.Position[2] = 0
	}

	if a.HasYaw() {
		io.Uint16(&a.Yaw)
	} else {
		a.Yaw = 0
	}

	if a.HasPitch() {
		io.Uint16(&a.Pitch)
	} else {
		a.Pitch = 0
	}
}

func (a *EntityDeltaMove) Play(ctx *PlayContext) {
	prevPos, ok := ctx.Playback().EntityPosition(ctx.Tx(), a.EntityID)
	prevRot, ok2 := ctx.Playback().EntityRotation(ctx.Tx(), a.EntityID)
	if !ok || !ok2 {
		// Entity doesn't exist, can't apply delta movement
		return
	}

	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().MoveEntity(ctx.Tx(), a.EntityID, prevPos, prevRot)
	})
	pos := vec32To64(a.Position)
	rot := DecodeRotation16(a.Yaw, a.Pitch)
	if !a.HasX() {
		pos[0] = prevPos[0]
	}
	if !a.HasY() {
		pos[1] = prevPos[1]
	}
	if !a.HasZ() {
		pos[2] = prevPos[2]
	}
	if !a.HasYaw() {
		rot[0] = prevRot[0]
	}
	if !a.HasPitch() {
		rot[1] = prevRot[1]
	}
	ctx.Playback().MoveEntity(ctx.Tx(), a.EntityID, pos, rot)
}
