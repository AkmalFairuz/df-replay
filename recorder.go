package replay

import (
	"bytes"
	"encoding/binary"
	"github.com/akmalfairuz/df-replay/action"
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/block"
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

	enableEntityMovementRecording bool
}

// NewRecorder creates a new recorder, returning a pointer to the recorder.
func NewRecorder(id uuid.UUID) *Recorder {
	return newRecorder(id, true)
}

// NewRecorderWithoutEntityMovementRecording creates a new recorder that does not record entity movements.
func NewRecorderWithoutEntityMovementRecording(id uuid.UUID) *Recorder {
	return newRecorder(id, false)
}

func newRecorder(id uuid.UUID, enableEntityMovementRecording bool) *Recorder {
	return &Recorder{
		id:                            id,
		nextID:                        1,
		buffer:                        bytes.NewBuffer(make([]byte, 0, 8192)), // 8KB
		pendingActions:                make(map[uint32][]action.Action, 6000), // 5 minutes
		closing:                       make(chan struct{}),
		playerIDs:                     make(map[uuid.UUID]uint32, 32),
		entityIDs:                     make(map[uuid.UUID]uint32, 32),
		lastPushedPlayerMovements:     make(map[uuid.UUID]mgl64.Vec3, 32),
		lastPushedEntityMovements:     make(map[uuid.UUID]mgl64.Vec3, 32),
		tick:                          1,
		enableEntityMovementRecording: enableEntityMovementRecording,
	}
}

// StartTicking ...
func (r *Recorder) StartTicking(w *world.World) {
	if r.enableEntityMovementRecording {
		r.entityMovementRecorder = newWorldEntityMovementRecorder(r)
	}

	r.mu.Lock()
	if r.w != nil {
		panic("recorder already started")
	}
	r.w = w
	r.mu.Unlock()

	if r.enableEntityMovementRecording {
		r.recording.Add(2)
		go r.entityMovementRecorder.StartTicking()
	} else {
		r.recording.Add(1)
	}
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
	case entity.TNTType:
		extraData["Fuse"] = int32(e.(*entity.Ent).Behaviour().(*entity.PassiveBehaviour).Fuse().Milliseconds())
	case entity.FallingBlockType:
		extraData["Block"] = int32(internal.BlockToHash(e.(*entity.Ent).Behaviour().(*entity.FallingBlockBehaviour).Block()))
	case entity.FireworkType:
		fw := e.(*entity.Ent).Behaviour().(*entity.FireworkBehaviour)
		extraData["Item"] = fw.Firework().EncodeNBT()
	case entity.ArrowType, entity.SplashPotionType:
		proj := e.(*entity.Ent).Behaviour().(*entity.ProjectileBehaviour)
		if potionId := proj.Potion().Uint8(); potionId != 0 {
			extraData["PotionID"] = int8(potionId)
		}
	}
	type hasOwner interface {
		Owner() *world.EntityHandle
	}
	if ent, ok := e.(*entity.Ent); ok {
		if ownerable, ok := ent.Behaviour().(hasOwner); ok {
			if owner := ownerable.Owner(); owner != nil {
				ownerID := r.PlayerIDByHandle(owner)
				if ownerID != 0 {
					extraData["Owner"] = int32(ownerID)
				}
			}
		}
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

	r.removeLastEntityMovement(e)
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

	r.removeLastPlayerMovement(p)
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
	r.pushPlayerAnimate(p, action.PlayerAnimateSwing)
}

// PushPlayerEating ...
func (r *Recorder) PushPlayerEating(p *player.Player) {
	r.pushPlayerAnimate(p, action.PlayerAnimateEating)
}

// PushPlayerHurt ...
func (r *Recorder) PushPlayerHurt(p *player.Player) {
	r.pushPlayerAnimate(p, action.PlayerAnimateHurt)
}

// PushPlayerCriticalHit ...
func (r *Recorder) PushPlayerCriticalHit(p *player.Player) {
	r.pushPlayerAnimate(p, action.PlayerAnimateCriticalHit)
}

// PushPlayerEnchantedHit ...
func (r *Recorder) PushPlayerEnchantedHit(p *player.Player) {
	r.pushPlayerAnimate(p, action.PlayerAnimateEnchantedHit)
}

// pushPlayerAnimate ...
func (r *Recorder) pushPlayerAnimate(p *player.Player, animation uint8) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}
	r.PushAction(&action.PlayerAnimate{
		PlayerID:  playerID,
		Animation: animation,
	})
}

// pushPlayerState ...
func (r *Recorder) pushPlayerState(p *player.Player, stateType uint8, value bool) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.SetPlayerState{
		PlayerID: playerID,
		Type:     stateType,
		Value:    value,
	})
}

// PushPlayerUsingItem ...
func (r *Recorder) PushPlayerUsingItem(p *player.Player, usingItem bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeUsingItem, usingItem)
}

// PushPlayerSneaking ...
func (r *Recorder) PushPlayerSneaking(p *player.Player, sneaking bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeSneaking, sneaking)
}

// PushPlayerVisibility ...
func (r *Recorder) PushPlayerVisibility(p *player.Player, visibility bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeVisibility, visibility)
}

// PushPlayerSwimming ...
func (r *Recorder) PushPlayerSwimming(p *player.Player, swimming bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeSwimming, swimming)
}

// PushPlayerGliding ...
func (r *Recorder) PushPlayerGliding(p *player.Player, gliding bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeGliding, gliding)
}

// PushPlayerCrawling ...
func (r *Recorder) PushPlayerCrawling(p *player.Player, crawling bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeCrawling, crawling)
}

// PushPlayerOnFire ...
func (r *Recorder) PushPlayerOnFire(p *player.Player, onFire bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeOnFire, onFire)
}

// PushPlayerSprinting ...
func (r *Recorder) PushPlayerSprinting(p *player.Player, sprinting bool) {
	r.pushPlayerState(p, action.SetPlayerStateTypeSprinting, sprinting)
}

// PushBlockBreakingSound ...
func (r *Recorder) PushBlockBreakingSound(pos cube.Pos, b world.Block) {
	r.pushBlockSound(pos, b, action.BlockSoundTypeBreaking)
}

// PushBlockPlaceSound ...
func (r *Recorder) PushBlockPlaceSound(pos cube.Pos, b world.Block) {
	r.pushBlockSound(pos, b, action.BlockSoundTypePlace)
}

// pushBlockSound ...
func (r *Recorder) pushBlockSound(pos cube.Pos, b world.Block, t uint8) {
	r.PushAction(&action.BlockSound{
		Position: cubeToBlockPos(pos),
		Block:    action.FromBlock(b),
		Type:     t,
	})
}

// PushLiquidSound ...
func (r *Recorder) PushLiquidSound(pos cube.Pos, l world.Liquid, isFill bool) {
	_, isWater := l.(block.Water)
	r.PushAction(&action.LiquidSound{
		Position: cubeToBlockPos(pos),
		IsWater:  isWater,
		Type:     lo.If(isFill, action.LiquidSoundTypeFill).Else(action.LiquidSoundTypeEmpty),
	})
}

// PushGeneralSound ...
func (r *Recorder) PushGeneralSound(pos mgl64.Vec3, s world.Sound) bool {
	soundID, ok := internal.ToSoundID(s)
	if !ok {
		return false
	}
	r.PushAction(&action.GeneralSound{
		Position: vec64To32(pos),
		SoundID:  soundID,
	})
	return true
}

// PushBlockBreakParticle ...
func (r *Recorder) PushBlockBreakParticle(pos cube.Pos, b world.Block) {
	r.pushBlockParticle(pos, b, action.BlockParticleTypeBreak, 0)
}

// PushBlockPunchingParticle ...
func (r *Recorder) PushBlockPunchingParticle(pos cube.Pos, b world.Block, face cube.Face) {
	r.pushBlockParticle(pos, b, action.BlockParticleTypePunching, uint8(face))
}

// PushGeneralParticle ...
func (r *Recorder) PushGeneralParticle(pos mgl64.Vec3, p world.Particle) bool {
	particleId, ok := internal.ToParticleID(p)
	if !ok {
		return false
	}
	r.PushAction(&action.GeneralParticle{
		Position:   vec64To32(pos),
		ParticleID: particleId,
	})
	return true
}

// pushBlockParticle ...
func (r *Recorder) pushBlockParticle(pos cube.Pos, b world.Block, t uint8, face uint8) {
	r.PushAction(&action.BlockParticle{
		Position: cubeToBlockPos(pos),
		Block:    action.FromBlock(b),
		Type:     t,
		Face:     face,
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

// PushPlayerArmorChange ...
func (r *Recorder) PushPlayerArmorChange(p *player.Player) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.PlayerArmorChange{
		PlayerID:   playerID,
		Helmet:     action.ItemFromStack(p.Armour().Helmet()),
		Chestplate: action.ItemFromStack(p.Armour().Chestplate()),
		Leggings:   action.ItemFromStack(p.Armour().Leggings()),
		Boots:      action.ItemFromStack(p.Armour().Boots()),
	})
}

// PushEmote ...
func (r *Recorder) PushEmote(p *player.Player, emote uuid.UUID) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}

	r.PushAction(&action.Emote{
		PlayerID: playerID,
		EmoteID:  emote,
	})
}

// PushStartCrackBlock ...
func (r *Recorder) PushStartCrackBlock(pos cube.Pos, dur time.Duration) {
	r.pushCrackBlock(pos, action.CrackBlockTypeStart, dur)
}

// PushStopCrackBlock ...
func (r *Recorder) PushStopCrackBlock(pos cube.Pos) {
	r.pushCrackBlock(pos, action.CrackBlockTypeStop, 0)
}

// PushContinueCrackBlock ...
func (r *Recorder) PushContinueCrackBlock(pos cube.Pos, dur time.Duration) {
	r.pushCrackBlock(pos, action.CrackBlockTypeContinue, dur)
}

// pushCrackBlock ...
func (r *Recorder) pushCrackBlock(pos cube.Pos, t uint8, dur time.Duration) {
	r.PushAction(&action.CrackBlock{
		Position:     cubeToBlockPos(pos),
		Type:         t,
		DurationInMs: uint16(dur.Milliseconds()),
	})
}

// PushPlayerTotemUse ...
func (r *Recorder) PushPlayerTotemUse(p *player.Player) {
	r.pushPlayerAnimate(p, action.PlayerAnimateTotemUse)
}

// PushPlayerSetVisibleEffects ...
func (r *Recorder) PushPlayerSetVisibleEffects(p *player.Player, effectIds []int) {
	playerID := r.PlayerID(p)
	if playerID == 0 {
		return
	}
	r.PushAction(&action.SetPlayerVisibleEffects{
		PlayerID: playerID,
		Effects:  lo.Map(effectIds, func(e int, _ int) uint8 { return uint8(e) }),
	})
}

// PushEntityFireworkExplosion ...
func (r *Recorder) PushEntityFireworkExplosion(e world.Entity) {
	r.pushEntityAnimate(e, action.EntityAnimateFireworkExplosion)
}

// PushEntityArrowShake ...
func (r *Recorder) PushEntityArrowShake(e world.Entity) {
	r.pushEntityAnimate(e, action.EntityAnimateArrowShake)
}

// pushEntityAnimate ...
func (r *Recorder) pushEntityAnimate(e world.Entity, animation uint8) {
	entityID := r.EntityID(e)
	if entityID == 0 {
		return
	}
	r.PushAction(&action.EntityAnimate{
		EntityID:  entityID,
		Animation: animation,
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

// removeLastPlayerMovement removes the last player movement from the recorder.
func (r *Recorder) removeLastPlayerMovement(p *player.Player) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lastPushedPlayerMovements, p.UUID())
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
	buf := internal.BufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		internal.BufferPool.Put(buf)
	}()
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
