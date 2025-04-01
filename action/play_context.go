package action

import "github.com/df-mc/dragonfly/server/world"

type PlayContext struct {
	tx        *world.Tx
	playback  Playback
	onReverse func(ctx *PlayContext)
}

func NewPlayContext(tx *world.Tx, playback Playback) *PlayContext {
	return &PlayContext{tx: tx, playback: playback}
}

func (ctx *PlayContext) Playback() Playback {
	return ctx.playback
}

func (ctx *PlayContext) Tx() *world.Tx {
	return ctx.tx
}

func (ctx *PlayContext) OnReverse(h func(ctx *PlayContext)) {
	ctx.onReverse = h
}

func (ctx *PlayContext) ReverseHandler() (func(ctx *PlayContext), bool) {
	if ctx.onReverse == nil {
		return nil, false
	}
	return ctx.onReverse, true
}
