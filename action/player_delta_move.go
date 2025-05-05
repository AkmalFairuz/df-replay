package action

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerDeltaMoveHasXFlag = 1 << iota
	PlayerDeltaMoveHasYFlag
	PlayerDeltaMoveHasZFlag
	PlayerDeltaMoveHasYawFlag
	PlayerDeltaMoveHasPitchFlag
)

type PlayerDeltaMove struct {
	Flags      uint8
	PlayerID   uint32
	Position   mgl32.Vec3
	Yaw, Pitch uint16
}

func (*PlayerDeltaMove) ID() uint8 {
	return IDPlayerDeltaMove
}

func (a *PlayerDeltaMove) HasX() bool {
	return a.Flags&PlayerDeltaMoveHasXFlag != 0
}

func (a *PlayerDeltaMove) HasY() bool {
	return a.Flags&PlayerDeltaMoveHasYFlag != 0
}

func (a *PlayerDeltaMove) HasZ() bool {
	return a.Flags&PlayerDeltaMoveHasZFlag != 0
}

func (a *PlayerDeltaMove) HasYaw() bool {
	return a.Flags&PlayerDeltaMoveHasYawFlag != 0
}

func (a *PlayerDeltaMove) HasPitch() bool {
	return a.Flags&PlayerDeltaMoveHasPitchFlag != 0
}

func (a *PlayerDeltaMove) Marshal(io protocol.IO) {
	io.Uint8(&a.Flags)
	io.Varuint32(&a.PlayerID)
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

func (a *PlayerDeltaMove) Play(ctx *PlayContext) {
	prevPos, ok := ctx.Playback().PlayerPosition(ctx.Tx(), a.PlayerID)
	prevRot, ok2 := ctx.Playback().PlayerRotation(ctx.Tx(), a.PlayerID)
	if !ok || !ok2 {
		// PANIC???
		return
	}

	ctx.OnReverse(func(ctx *PlayContext) {
		ctx.Playback().MovePlayer(ctx.Tx(), a.PlayerID, prevPos, prevRot)
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
	ctx.Playback().MovePlayer(ctx.Tx(), a.PlayerID, pos, rot)
}
