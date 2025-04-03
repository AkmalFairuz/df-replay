package replay

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"time"
)

type movementData struct {
	Pos mgl64.Vec3
	Rot cube.Rotation
}

type WorldEntityMovementRecorder struct {
	r *Recorder

	lastMovement map[uuid.UUID]movementData
}

// newWorldEntityMovementRecorder ...
func newWorldEntityMovementRecorder(r *Recorder) *WorldEntityMovementRecorder {
	return &WorldEntityMovementRecorder{
		r:            r,
		lastMovement: make(map[uuid.UUID]movementData, 128),
	}
}

// StartTicking ...
func (r *WorldEntityMovementRecorder) StartTicking() {
	ticker := time.NewTicker(time.Second / 20)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			<-r.r.w.Exec(r.Tick)
		case <-r.r.closing:
			r.r.recording.Done()
			return
		}
	}
}

// Tick ...
func (r *WorldEntityMovementRecorder) Tick(tx *world.Tx) {
	for e := range tx.Entities() {
		lastMovement, ok := r.lastMovement[e.H().UUID()]
		if !ok {
			r.r.PushEntityMovement(e, e.Position(), e.Rotation())
			r.lastMovement[e.H().UUID()] = movementData{
				Pos: e.Position(),
				Rot: e.Rotation(),
			}
			continue
		}
		if lastMovement.Pos != e.Position() || lastMovement.Rot != e.Rotation() {
			r.r.PushEntityMovement(e, e.Position(), e.Rotation())
			r.lastMovement[e.H().UUID()] = movementData{
				Pos: e.Position(),
				Rot: e.Rotation(),
			}
		}
	}
}
