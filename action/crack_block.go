package action

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"time"
)

const (
	CrackBlockTypeStart uint8 = iota
	CrackBlockTypeContinue
	CrackBlockTypeStop
)

type CrackBlock struct {
	Position     protocol.BlockPos
	Type         uint8
	DurationInMs uint16
}

func (a *CrackBlock) ID() uint8 {
	return IDCrackBlock
}

func (a *CrackBlock) Marshal(io protocol.IO) {
	io.BlockPos(&a.Position)
	io.Uint8(&a.Type)
	if a.Type != CrackBlockTypeStop {
		io.Uint16(&a.DurationInMs)
	}
}

func (a *CrackBlock) Play(ctx *PlayContext) {
	dur := time.Millisecond * time.Duration(a.DurationInMs)
	ctx.OnReverse(func(ctx *PlayContext) {
		switch a.Type {
		case CrackBlockTypeStart:
			ctx.Playback().StopCrackBlock(ctx.Tx(), blockPosToCubePos(a.Position))
		default:
			// TODO:
		}
	})
	switch a.Type {
	case CrackBlockTypeStart:
		ctx.Playback().StartCrackBlock(ctx.Tx(), blockPosToCubePos(a.Position), dur)
	case CrackBlockTypeContinue:
		ctx.Playback().ContinueCrackBlock(ctx.Tx(), blockPosToCubePos(a.Position), dur)
	case CrackBlockTypeStop:
		ctx.Playback().StopCrackBlock(ctx.Tx(), blockPosToCubePos(a.Position))
	}
}
