package replay

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"time"
)

type movementData struct {
	Pos mgl64.Vec3
	Rot cube.Rotation
}

func (m movementData) Equal(other movementData) bool {
	const threshold = 0.001
	return m.Pos.ApproxEqualThreshold(other.Pos, threshold) &&
		mgl64.FloatEqualThreshold(m.Rot[0], other.Rot[0], threshold) &&
		mgl64.FloatEqualThreshold(m.Rot[1], other.Rot[1], threshold)
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
			select {
			case <-r.r.closing:
				r.r.recording.Done()
				return
			case <-r.r.w.Exec(r.Tick):
			}
		case <-r.r.closing:
			r.r.recording.Done()
			return
		}
	}
}

// Tick ...
func (r *WorldEntityMovementRecorder) Tick(tx *world.Tx) {
	select {
	case <-r.r.closing:
		return
	default:
	}
	for e := range tx.Entities() {
		if _, isPlayer := e.(*player.Player); isPlayer {
			continue
		}

		movData := movementData{
			Pos: e.Position(),
			Rot: e.Rotation(),
		}

		lastMovement, ok := r.lastMovement[e.H().UUID()]
		if !ok {
			r.r.PushEntityMovement(e, e.Position(), e.Rotation())
			r.lastMovement[e.H().UUID()] = movData
			continue
		}
		if !lastMovement.Equal(movData) {
			r.r.PushEntityMovement(e, e.Position(), e.Rotation())
			r.lastMovement[e.H().UUID()] = movData
		}
	}
}
