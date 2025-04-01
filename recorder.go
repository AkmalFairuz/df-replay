package replay

import (
	"bytes"
	"encoding/binary"
	"github.com/akmalfairuz/df-replay/action"
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
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
	playerIDs      map[uuid.UUID]uint32
	nextID         uint32
}

// NewRecorder creates a new recorder, returning a pointer to the recorder.
func NewRecorder(id uuid.UUID) *Recorder {
	return &Recorder{
		id:             id,
		buffer:         bytes.NewBuffer(make([]byte, 0, 8192)), // 8KB
		pendingActions: make(map[uint32][]action.Action, 512),
	}
}

// AddPlayer ...
func (r *Recorder) AddPlayer(p *player.Player) {
	addedBefore := false
	r.mu.Lock()
	var playerID uint32
	if id, ok := r.playerIDs[p.UUID()]; ok {
		playerID = id
		addedBefore = true
	} else {
		playerID = r.nextID
		r.playerIDs[p.UUID()] = playerID
		r.nextID++
	}
	r.nextID++
	r.mu.Unlock()

	if !addedBefore {
		sk := p.Skin()
		r.PushAction(skinToAction(playerID, sk))
	}

	mainHand, offHand := p.HeldItems()

	r.PushAction(&action.PlayerSpawn{
		PlayerID:   playerID,
		PlayerName: p.Name(),
		Position:   vec64To32(p.Position()),
		Yaw:        float32(p.Rotation().Yaw()),
		Pitch:      float32(p.Rotation().Pitch()),
		Helmet:     action.ItemFromStack(p.Armour().Helmet()),
		Chestplate: action.ItemFromStack(p.Armour().Chestplate()),
		Leggings:   action.ItemFromStack(p.Armour().Leggings()),
		Boots:      action.ItemFromStack(p.Armour().Boots()),
		MainHand:   action.ItemFromStack(mainHand),
		OffHand:    action.ItemFromStack(offHand),
	})
}

// playerID ...
func (r *Recorder) playerID(p *player.Player) uint32 {
	r.mu.Lock()
	playerID, ok := r.playerIDs[p.UUID()]
	r.mu.Unlock()

	if !ok {
		return 0
	}
	return playerID
}

// RemovePlayer ...
func (r *Recorder) RemovePlayer(p *player.Player) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerDespawn{
		PlayerID: playerID,
	})
}

// PushPlayerMovement ...
func (r *Recorder) PushPlayerMovement(p *player.Player, pos mgl64.Vec3, rot cube.Rotation) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerMove{
		PlayerID: playerID,
		Position: vec64To32(pos),
		Yaw:      float32(rot.Yaw()),
		Pitch:    float32(rot.Pitch()),
	})
}

// PushPlayerHandChange ...
func (r *Recorder) PushPlayerHandChange(p *player.Player, mainHand, offHand item.Stack) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerHandChange{
		PlayerID: playerID,
		MainHand: action.ItemFromStack(mainHand),
		OffHand:  action.ItemFromStack(offHand),
	})
}

// PushPlayerSwingArm ...
func (r *Recorder) PushPlayerSwingArm(p *player.Player) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerAnimate{
		PlayerID:  playerID,
		Animation: action.PlayerAnimateSwing,
	})
}

// PushPlayerUsingItem ...
func (r *Recorder) PushPlayerUsingItem(p *player.Player, usingItem bool) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	var animation uint8
	if usingItem {
		animation = action.PlayerAnimateStartUsingItem
	} else {
		animation = action.PlayerAnimateStopUsingItem
	}

	r.PushAction(&action.PlayerAnimate{
		PlayerID:  playerID,
		Animation: animation,
	})
}

// PushPlayerEating ...
func (r *Recorder) PushPlayerEating(p *player.Player) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerAnimate{
		PlayerID:  playerID,
		Animation: action.PlayerAnimateEating,
	})
}

// PushPlayerSneaking ...
func (r *Recorder) PushPlayerSneaking(p *player.Player, sneaking bool) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	var animation uint8
	if sneaking {
		animation = action.PlayerAnimateSneak
	} else {
		animation = action.PlayerAnimateStopSneak
	}

	r.PushAction(&action.PlayerAnimate{
		PlayerID:  playerID,
		Animation: animation,
	})
}

// PushPlayerHurt ...
func (r *Recorder) PushPlayerHurt(p *player.Player) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerAnimate{
		PlayerID:  playerID,
		Animation: action.PlayerAnimateHurt,
	})
}

// PushPlaceBlock ...
func (r *Recorder) PushPlaceBlock(pos cube.Pos, b world.Block) {
	r.PushAction(&action.PlaceBlock{
		Position:  cubeToBlockPos(pos),
		BlockHash: internal.BlockToHash(b),
	})
}

// PushBreakBlock ...
func (r *Recorder) PushBreakBlock(pos cube.Pos) {
	r.PushAction(&action.BreakBlock{
		Position: cubeToBlockPos(pos),
	})
}

// PushSetBlock ...
func (r *Recorder) PushSetBlock(pos cube.Pos, b world.Block) {
	r.PushAction(&action.SetBlock{
		Position:  cubeToBlockPos(pos),
		BlockHash: internal.BlockToHash(b),
		Layer:     0, // TODO
	})
}

// PushSkinChange ...
func (r *Recorder) PushSkinChange(p *player.Player, sk skin.Skin) {
	playerID := r.playerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(skinToAction(playerID, sk))
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
