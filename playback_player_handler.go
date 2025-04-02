package replay

import "github.com/df-mc/dragonfly/server/player"

type PlaybackPlayerHandler struct {
	player.NopHandler

	w *Playback
}

// NewPlaybackPlayerHandler ...
func NewPlaybackPlayerHandler(w *Playback) *PlaybackPlayerHandler {
	return &PlaybackPlayerHandler{
		w: w,
	}
}
