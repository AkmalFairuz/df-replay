package replay

import (
	"bytes"
	"encoding/binary"
	"github.com/akmalfairuz/df-replay/action"
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
	"sync"
	"time"
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
	entityIDs      map[uuid.UUID]uint32
	nextID         uint32

	w *world.World

	recording sync.WaitGroup
	closing   chan struct{}
	once      sync.Once

	lastPushedPlayerMovements map[uuid.UUID]mgl64.Vec3
	lastPushedEntityMovements map[uuid.UUID]mgl64.Vec3

	entityMovementRecorder *WorldEntityMovementRecorder
}

// NewRecorder creates a new recorder, returning a pointer to the recorder.
func NewRecorder(id uuid.UUID) *Recorder {
	return &Recorder{
		id:                        id,
		nextID:                    1,
		buffer:                    bytes.NewBuffer(make([]byte, 0, 8192)), // 8KB
		pendingActions:            make(map[uint32][]action.Action, 6000), // 5 minutes
		closing:                   make(chan struct{}),
		playerIDs:                 make(map[uuid.UUID]uint32, 32),
		entityIDs:                 make(map[uuid.UUID]uint32, 32),
		lastPushedPlayerMovements: make(map[uuid.UUID]mgl64.Vec3, 32),
		lastPushedEntityMovements: make(map[uuid.UUID]mgl64.Vec3, 32),
		tick:                      1,
	}
}

// StartTicking ...
func (r *Recorder) StartTicking(w *world.World) {
	r.entityMovementRecorder = newWorldEntityMovementRecorder(r)

	r.mu.Lock()
	if r.w != nil {
		panic("recorder already started")
	}
	r.w = w
	r.mu.Unlock()

	r.recording.Add(2)

	go r.entityMovementRecorder.StartTicking()
	go r.startTickCounter()
}

// startTickCounter ...
func (r *Recorder) startTickCounter() {
	ticker := time.NewTicker(time.Second / 20)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			r.tick++
			r.mu.Unlock()

			if r.tick%3600 == 0 { // 3 minutes
				r.Flush()
			}
		case <-r.closing:
			r.recording.Done()
			return
		}
	}
}

// CloseAndSaveActions ...
func (r *Recorder) CloseAndSaveActions(w io.Writer) error {
	var err error
	r.once.Do(func() {
		err = r.doCloseAndSaveActions(w)
	})
	return err
}

// doClose ...
func (r *Recorder) doCloseAndSaveActions(w io.Writer) error {
	close(r.closing)
	r.recording.Wait()
	r.doFlush(true)
	if err := r.saveActions(w); err != nil {
		return err
	}
	r.buffer = nil
	r.w = nil
	return nil
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
		NameTag:    p.NameTag(),
		Position:   vec64To32(p.Position()),
		Yaw:        action.EncodeYaw16(float32(p.Rotation().Yaw())),
		Pitch:      action.EncodePitch16(float32(p.Rotation().Pitch())),
		Helmet:     action.ItemFromStack(p.Armour().Helmet()),
		Chestplate: action.ItemFromStack(p.Armour().Chestplate()),
		Leggings:   action.ItemFromStack(p.Armour().Leggings()),
		Boots:      action.ItemFromStack(p.Armour().Boots()),
		MainHand:   action.ItemFromStack(mainHand),
		OffHand:    action.ItemFromStack(offHand),
	})
}

// AddEntity ...
func (r *Recorder) AddEntity(e world.Entity) {
	r.mu.Lock()
	var entityID uint32
	if id, ok := r.entityIDs[e.H().UUID()]; ok {
		entityID = id
	} else {
		entityID = r.nextID
		r.entityIDs[e.H().UUID()] = entityID
		r.nextID++
	}
	r.mu.Unlock()
	identifier := e.H().Type().EncodeEntity()
	if v, ok := e.(session.NetworkEncodeableEntity); ok {
		identifier = v.NetworkEncodeEntity()
	}
	extraData := make(map[string]any)
	switch e.H().Type() {
	case entity.ItemType:
		stack := e.(*entity.Ent).Behaviour().(*entity.ItemBehaviour).Item()
		extraData["Item"] = int64(internal.ItemToHash(stack.Item()))
		extraData["ItemCount"] = int32(stack.Count())
	case entity.TextType:
		extraData["IsTextType"] = byte(1)
	}
	var nameTag string
	if ent, ok := e.(interface{ NameTag() string }); ok {
		nameTag = ent.NameTag()
	}
	r.PushAction(&action.EntitySpawn{
		EntityID:         entityID,
		EntityIdentifier: identifier,
		NameTag:          nameTag,
		Position:         vec64To32(e.Position()),
		Yaw:              action.EncodeYaw16(float32(e.Rotation().Yaw())),
		Pitch:            action.EncodePitch16(float32(e.Rotation().Pitch())),
		ExtraData:        extraData,
	})
}

// RemoveEntity ...
func (r *Recorder) RemoveEntity(e world.Entity) {
	entityID := r.EntityID(e)
	if entityID == 0 {
		return
	}

	r.PushAction(&action.EntityDespawn{
		EntityID: entityID,
	})
}

// PlayerID ...
func (r *Recorder) PlayerID(p *player.Player) uint32 {
	r.mu.Lock()
	playerID, ok := r.playerIDs[p.UUID()]
	r.mu.Unlock()

	if !ok {
		return 0
	}
	return playerID
}

// PlayerIDByHandle ...
func (r *Recorder) PlayerIDByHandle(h *world.EntityHandle) uint32 {
	r.mu.Lock()
	playerID, ok := r.playerIDs[h.UUID()]
	r.mu.Unlock()

	if !ok {
		return 0
	}
	return playerID
}

// EntityID ...
func (r *Recorder) EntityID(e world.Entity) uint32 {
	r.mu.Lock()
	entityID, ok := r.entityIDs[e.H().UUID()]
	r.mu.Unlock()

	if !ok {
		return 0
	}
	return entityID
}

// EntityIDByHandle ...
func (r *Recorder) EntityIDByHandle(h *world.EntityHandle) uint32 {
	r.mu.Lock()
	entityID, ok := r.entityIDs[h.UUID()]
	r.mu.Unlock()

	if !ok {
		return 0
	}
	return entityID
}

// RemovePlayer ...
func (r *Recorder) RemovePlayer(p *player.Player) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerDespawn{
		PlayerID: playerID,
	})
}

// PushPlayerMovement ...
func (r *Recorder) PushPlayerMovement(p *player.Player, pos mgl64.Vec3, rot cube.Rotation) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if lastPos, ok := r.lastPushedPlayerMovements[p.UUID()]; ok {
		flags := uint8(0)
		var changedPos mgl32.Vec3
		var yaw, pitch uint16
		if !mgl64.FloatEqual(pos[0], lastPos[0]) {
			flags |= action.PlayerDeltaMoveHasXFlag
			changedPos[0] = float32(pos[0])
		}
		if !mgl64.FloatEqual(pos[1], lastPos[1]) {
			flags |= action.PlayerDeltaMoveHasYFlag
			changedPos[1] = float32(pos[1])
		}
		if !mgl64.FloatEqual(pos[2], lastPos[2]) {
			flags |= action.PlayerDeltaMoveHasZFlag
			changedPos[2] = float32(pos[2])
		}

		prevRot := p.Rotation()
		if !mgl64.FloatEqual(rot[0], prevRot[0]) {
			flags |= action.PlayerDeltaMoveHasYawFlag
			yaw = action.EncodeYaw16(float32(rot[0]))
		}
		if !mgl64.FloatEqual(rot[1], prevRot[1]) {
			flags |= action.PlayerDeltaMoveHasPitchFlag
			pitch = action.EncodePitch16(float32(rot[1]))
		}

		r.pushActionNoMutex(&action.PlayerDeltaMove{
			Flags:    flags,
			PlayerID: playerID,
			Position: changedPos,
			Yaw:      yaw,
			Pitch:    pitch,
		})
	} else {
		r.pushActionNoMutex(&action.PlayerMove{
			PlayerID: playerID,
			Position: vec64To32(pos),
			Yaw:      action.EncodeYaw16(float32(rot[0])),
			Pitch:    action.EncodePitch16(float32(rot[1])),
		})
	}
	r.lastPushedPlayerMovements[p.UUID()] = pos
}

// PushEntityMovement ...
func (r *Recorder) PushEntityMovement(e world.Entity, pos mgl64.Vec3, rot cube.Rotation) {
	entityID := r.EntityID(e)
	if entityID == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if lastPos, ok := r.lastPushedEntityMovements[e.H().UUID()]; ok {
		flags := uint8(0)
		var changedPos mgl32.Vec3
		var yaw, pitch uint16
		if pos[0] != lastPos[0] {
			flags |= action.EntityDeltaMoveHasXFlag
			changedPos[0] = float32(pos[0])
		}
		if pos[1] != lastPos[1] {
			flags |= action.EntityDeltaMoveHasYFlag
			changedPos[1] = float32(pos[1])
		}
		if pos[2] != lastPos[2] {
			flags |= action.EntityDeltaMoveHasZFlag
			changedPos[2] = float32(pos[2])
		}

		prevRot := e.Rotation()
		if rot[0] != prevRot[0] {
			flags |= action.EntityDeltaMoveHasYawFlag
			yaw = action.EncodeYaw16(float32(rot[0]))
		}
		if rot[1] != prevRot[1] {
			flags |= action.EntityDeltaMoveHasPitchFlag
			pitch = action.EncodePitch16(float32(rot[1]))
		}

		r.pushActionNoMutex(&action.EntityDeltaMove{
			Flags:    flags,
			EntityID: entityID,
			Position: changedPos,
			Yaw:      yaw,
			Pitch:    pitch,
		})
	} else {
		r.pushActionNoMutex(&action.EntityMove{
			EntityID: entityID,
			Position: vec64To32(pos),
			Yaw:      action.EncodeYaw16(float32(rot[0])),
			Pitch:    action.EncodePitch16(float32(rot[1])),
		})
	}
	r.lastPushedEntityMovements[e.H().UUID()] = pos
}

// PushPlayerHandChange ...
func (r *Recorder) PushPlayerHandChange(p *player.Player, mainHand, offHand item.Stack) {
	playerID := r.PlayerID(p)
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
	playerID := r.PlayerID(p)
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
	playerID := r.PlayerID(p)
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
	playerID := r.PlayerID(p)
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
	playerID := r.PlayerID(p)
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
	playerID := r.PlayerID(p)
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
		Position: cubeToBlockPos(pos),
		Block:    action.FromBlock(b),
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
		Position: cubeToBlockPos(pos),
		Block:    action.FromBlock(b),
	})
}

// PushSkinChange ...
func (r *Recorder) PushSkinChange(p *player.Player, sk skin.Skin) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(skinToAction(playerID, sk))
}

// PushSetLiquid ...
func (r *Recorder) PushSetLiquid(pos cube.Pos, l world.Liquid) {
	r.PushAction(&action.SetLiquid{
		Position:   cubeToBlockPos(pos),
		LiquidHash: internal.BlockToHash(l),
	})
}

// PushSetPlayerNameTag ...
func (r *Recorder) PushSetPlayerNameTag(p *player.Player, nameTag string) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerNameTagUpdate{
		PlayerID: playerID,
		NameTag:  nameTag,
	})
}

// PushSetEntityNameTag ...
func (r *Recorder) PushSetEntityNameTag(e world.Entity, nameTag string) {
	entityID := r.EntityID(e)
	if entityID == 0 {
		return
	}

	r.PushAction(&action.EntityNameTagUpdate{
		EntityID: entityID,
		NameTag:  nameTag,
	})
}

// PushChestUpdate ...
func (r *Recorder) PushChestUpdate(pos cube.Pos, open bool) {
	r.PushAction(&action.ChestUpdate{
		Position: cubeToBlockPos(pos),
		Open:     open,
	})
}

// EncodeItem ...
func (r *Recorder) EncodeItem(s item.Stack) action.Item {
	return action.ItemFromStack(s)
}

// PushAction pushes an action to the recorder, to be written to the buffer at a later time.
func (r *Recorder) PushAction(a action.Action) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pushActionNoMutex(a)
}

// removeLastEntityMovement removes the last entity movement from the recorder.
func (r *Recorder) removeLastEntityMovement(e world.Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lastPushedEntityMovements, e.H().UUID())
}

// pushActionNoMutex pushes an action to the recorder without locking the mutex.
func (r *Recorder) pushActionNoMutex(a action.Action) {
	if _, ok := r.pendingActions[r.tick]; !ok {
		r.pendingActions[r.tick] = make([]action.Action, 0, 4)
	}
	r.pendingActions[r.tick] = append(r.pendingActions[r.tick], a)
}

// Flush flushes all pending actions to the buffer, writing them to the buffer.
func (r *Recorder) Flush() {
	select {
	case <-r.closing:
		return
	default:
	}

	r.doFlush(false)
}

func (r *Recorder) doFlush(closing bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	w := protocol.NewWriter(r.buffer, 0)
	untilTick := r.tick
	if !closing {
		untilTick--
	}
	for tick := r.flushedTick + 1; tick <= untilTick; tick++ {
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
	r.flushedTick = untilTick
}

// saveActions ...
func (r *Recorder) saveActions(w io.Writer) error {
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
	defer func() {
		_ = encoder.Close()
	}()

	if _, err := encoder.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
