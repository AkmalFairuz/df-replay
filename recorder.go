package replay

import (
	"bytes"
	"encoding/binary"
	"github.com/akmalfairuz/df-replay/action"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
	"sync"
)

// Recorder ...
type Recorder struct {
	id uuid.UUID
	mu sync.Mutex

	buffer         *bytes.Buffer
	bufferTickLen  int
	pendingActions map[uint32][]action.Action
	tick           uint32
	flushedTick    uint32
}

// NewRecorder creates a new recorder, returning a pointer to the recorder.
func NewRecorder(id uuid.UUID) *Recorder {
	return &Recorder{
		id:             id,
		buffer:         bytes.NewBuffer(make([]byte, 0, 8192)), // 8KB
		pendingActions: make(map[uint32][]action.Action, 512),
	}
}

// PushAction pushes an action to the recorder, to be written to the buffer at a later time.
func (r *Recorder) PushAction(a action.Action) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.pendingActions[r.tick]; !ok {
		r.pendingActions[r.tick] = make([]action.Action, 0, 4)
	}
	r.pendingActions[r.tick] = append(r.pendingActions[r.tick], a)
}

// Flush flushes all pending actions to the buffer, writing them to the buffer.
func (r *Recorder) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()

	w := protocol.NewWriter(r.buffer, 0)
	for tick := r.flushedTick + 1; tick < r.tick; tick++ {
		r.bufferTickLen++
		w.Varuint32(lo.ToPtr(tick))
		if actions, ok := r.pendingActions[tick]; ok {
			w.Varuint32(lo.ToPtr(uint32(len(actions))))
			for i, a := range actions {
				action.Write(w, a)
				// Set to nil to improve GC performance.
				r.pendingActions[tick][i] = nil
			}
			delete(r.pendingActions, tick)
		} else {
			w.Varuint32(lo.ToPtr(uint32(0)))
		}
	}
	r.flushedTick = r.tick - 1
}

// SaveActions ...
func (r *Recorder) SaveActions(w io.Writer) error {
	buf := bytes.NewBuffer(nil)
	r.mu.Lock()
	if err := binary.Write(buf, binary.LittleEndian, uint32(r.bufferTickLen)); err != nil {
		return err
	}
	if _, err := io.Copy(buf, r.buffer); err != nil {
		r.mu.Unlock()
		return err
	}
	r.mu.Unlock()

	encoder, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return err
	}
	defer encoder.Close()

	if _, err := encoder.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
