package replay

import (
	"github.com/akmalfairuz/df-replay/action"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"iter"
	"sync"
	"sync/atomic"
	"time"
)

// Playback handles the playback of recorded actions.
type Playback struct {
	w    *world.World
	data *Data

	playbackTick uint
	stopped      bool
	paused       bool
	speed        float64
	reverse      bool
	ended        bool

	closed  atomic.Bool
	running sync.WaitGroup
	once    sync.Once
	closing chan struct{}

	players         map[uint32]*Player
	entities        map[uint32]*Entity
	skins           map[uint32]skin.Skin
	reverseHandlers map[uint32][]func(ctx *action.PlayContext)
	chestState      map[cube.Pos]bool
}

// Compile time check to ensure that Playback implements action.Playback.
var _ action.Playback = (*Playback)(nil)

// NewPlayback creates a new playback instance with the given world and data.
func NewPlayback(w *world.World, data *Data) *Playback {
	return &Playback{
		w:               w,
		data:            data,
		players:         make(map[uint32]*Player),
		entities:        make(map[uint32]*Entity),
		skins:           make(map[uint32]skin.Skin),
		reverseHandlers: make(map[uint32][]func(ctx *action.PlayContext)),
		closing:         make(chan struct{}),
		speed:           1.0,
		chestState:      make(map[cube.Pos]bool, 16),
	}
}

// Play starts the playback.
func (w *Playback) Play() {
	w.running.Add(1)
	go w.doTicking()
}

func (w *Playback) doTicking() {
	ticker := time.NewTicker(time.Second / 100) // 100hz base tick
	defer ticker.Stop()

	tickAccumulator := 0.0

	for {
		select {
		case <-ticker.C:
			// target ticks per second
			tickRate := 20.0 * w.speed
			// because we are at 100hz
			tickAccumulator += tickRate / 100

			// if accumulator >= 1, run one or more ticks to catch up:
			for tickAccumulator >= 1.0 {
				tickAccumulator -= 1.0
				select {
				case <-w.closing:
					w.running.Done()
					return
				case <-w.w.Exec(w.Tick):
				}
			}

		case <-w.closing:
			w.running.Done()
			return
		}
	}
}

func (w *Playback) UpdateChestState(tx *world.Tx, pos cube.Pos, open bool) {
	w.chestState[pos] = open
	for _, v := range tx.Viewers(pos.Vec3Centre()) {
		if open {
			v.ViewBlockAction(pos, block.OpenAction{})
		} else {
			v.ViewBlockAction(pos, block.CloseAction{})
		}
	}
}

func (w *Playback) ChestState(_ *world.Tx, pos cube.Pos) bool {
	if open, ok := w.chestState[pos]; ok {
		return open
	}
	return false
}

func (w *Playback) Liquid(tx *world.Tx, pos cube.Pos) (world.Liquid, bool) {
	return tx.Liquid(pos)
}

func (w *Playback) SetLiquid(tx *world.Tx, pos cube.Pos, l world.Liquid) {
	tx.SetLiquid(pos, l)
}

func (w *Playback) SetPlayerNameTag(tx *world.Tx, id uint32, nameTag string) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.SetNameTag(nameTag)
}

func (w *Playback) SetEntityNameTag(tx *world.Tx, id uint32, nameTag string) {
	e, ok := w.openEntity(tx, id)
	if !ok {
		return
	}
	e.SetNameTag(nameTag)
}

func (w *Playback) PlayerNameTag(tx *world.Tx, id uint32) string {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return ""
	}
	return p.NameTag()
}

func (w *Playback) EntityNameTag(tx *world.Tx, id uint32) string {
	e, ok := w.openEntity(tx, id)
	if !ok {
		return ""
	}
	return e.NameTag()
}

func (w *Playback) SpawnEntity(tx *world.Tx, id uint32, identifier, nameTag string, pos mgl64.Vec3, rot cube.Rotation, extraData map[string]any) {
	opts := &world.EntitySpawnOpts{
		Position: pos,
		Rotation: rot,
		Velocity: mgl64.Vec3{},
		NameTag:  nameTag,
	}
	h := opts.New(entityType, entityBehaviourConfig{
		Identifier: identifier,
		ExtraData:  extraData,
	})
	tx.AddEntity(h)
	l := world.NewLoader(4, tx.World(), world.NopViewer{})
	l.Move(tx, pos)
	l.Load(tx, 4)
	w.entities[id] = &Entity{
		id:         id,
		h:          h,
		identifier: identifier,
		extraData:  extraData,
		l:          l,
	}
}

func (w *Playback) DespawnEntity(tx *world.Tx, id uint32) {
	ent, ok := w.openEntity(tx, id)
	if !ok {
		return
	}
	tx.RemoveEntity(ent)
	e, _ := w.entities[id]
	e.l.Close(tx)
	delete(w.entities, id)
}

func (w *Playback) MoveEntity(tx *world.Tx, id uint32, pos mgl64.Vec3, rot cube.Rotation) {
	ent, ok := w.openEntity(tx, id)
	if !ok {
		return
	}
	if v, ok := toAny(ent).(interface {
		SetPosAndRot(pos mgl64.Vec3, rot cube.Rotation)
	}); ok {
		v.SetPosAndRot(pos, rot)
	} else {
		updateEntEntityData(ent, "Pos", pos)
		updateEntEntityData(ent, "Rot", rot)
	}

	for _, v := range tx.Viewers(ent.Position()) {
		v.ViewEntityMovement(ent, pos, rot, false)
	}

	e, _ := w.entities[id]
	e.l.Move(tx, pos)
	e.l.Load(tx, 4)
}

func (w *Playback) EntityPosition(tx *world.Tx, id uint32) (mgl64.Vec3, bool) {
	ent, ok := w.openEntity(tx, id)
	if !ok {
		return mgl64.Vec3{}, false
	}
	return ent.Position(), true
}

func (w *Playback) EntityRotation(tx *world.Tx, id uint32) (cube.Rotation, bool) {
	ent, ok := w.openEntity(tx, id)
	if !ok {
		return cube.Rotation{}, false
	}
	return ent.Rotation(), true
}

func (w *Playback) EntityIdentifier(id uint32) (string, bool) {
	e, ok := w.entities[id]
	if !ok {
		return "", false
	}
	return e.identifier, true
}

func (w *Playback) EntityExtraData(id uint32) (map[string]any, bool) {
	e, ok := w.entities[id]
	if !ok {
		return nil, false
	}
	return e.extraData, true
}

func (w *Playback) PlayerSkin(id uint32) (skin.Skin, bool) {
	s, ok := w.skins[id]
	return s, ok
}

func (w *Playback) PlayerPosition(tx *world.Tx, id uint32) (mgl64.Vec3, bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return mgl64.Vec3{}, false
	}
	return p.Position(), true
}

func (w *Playback) PlayerRotation(tx *world.Tx, id uint32) (cube.Rotation, bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return cube.Rotation{}, false
	}
	return p.Rotation(), true
}

func (w *Playback) PlayerArmours(tx *world.Tx, id uint32) (helmet, chestplate, leggings, boots item.Stack, ok bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		ok = false
		return
	}
	armour := p.Armour()
	return armour.Helmet(), armour.Chestplate(), armour.Leggings(), armour.Boots(), true
}

func (w *Playback) PlayerHeldItems(tx *world.Tx, id uint32) (mainHand, offHand item.Stack, ok bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		ok = false
		return
	}
	mainHand, offHand = p.HeldItems()
	return mainHand, offHand, true
}

func (w *Playback) PlayerSneaking(tx *world.Tx, id uint32) bool {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return false
	}
	return p.Sneaking()
}

func (w *Playback) PlayerUsingItem(tx *world.Tx, id uint32) bool {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return false
	}
	return p.UsingItem()
}

func (w *Playback) UpdatePlayerArmours(tx *world.Tx, id uint32, helmet, chestplate, leggings, boots item.Stack) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.Armour().Set(helmet, chestplate, leggings, boots)
}

func (w *Playback) Block(tx *world.Tx, pos cube.Pos) world.Block {
	return tx.Block(pos)
}

func (w *Playback) MovePlayer(tx *world.Tx, id uint32, pos mgl64.Vec3, rot cube.Rotation) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.MoveSmooth(pos, rot)

	p2, _ := w.players[id]
	p2.l.Move(tx, pos)
	p2.l.Load(tx, 4)
}

func (w *Playback) SetBlock(tx *world.Tx, pos cube.Pos, b world.Block) {
	tx.SetBlock(pos, b, &world.SetOpts{
		DisableBlockUpdates:       true,
		DisableLiquidDisplacement: true,
	})
}

func (w *Playback) SpawnPlayer(tx *world.Tx, username, nameTag string, id uint32, pos mgl64.Vec3, rot cube.Rotation, armour [4]item.Stack, heldItems [2]item.Stack) {
	opts := &world.EntitySpawnOpts{
		Position: pos,
		Rotation: rot,
		NameTag:  username,
	}
	conf := player.Config{
		Name:      nameTag,
		Armour:    inventory.NewArmour(nil),
		Inventory: inventory.New(54, nil),
		OffHand:   inventory.New(1, nil),
	}
	conf.Position = pos
	conf.Rotation = rot
	conf.HeldSlot = 0
	conf.Armour.Set(armour[0], armour[1], armour[2], armour[3])
	_ = conf.Inventory.SetItem(0, heldItems[0])
	_ = conf.OffHand.SetItem(0, heldItems[1])
	if s, ok := w.skins[id]; ok {
		conf.Skin = s
	}
	h := opts.New(playerType, conf)
	l := world.NewLoader(4, tx.World(), world.NopViewer{})
	l.Move(tx, pos)
	l.Load(tx, 4)
	w.players[id] = &Player{
		id:   id,
		name: username,
		h:    h,
		l:    l,
	}
	tx.AddEntity(h)
}

func (w *Playback) PlayerName(id uint32) string {
	p, ok := w.players[id]
	if !ok {
		return ""
	}
	return p.name
}

func (w *Playback) PlayerArmourItems(tx *world.Tx, id uint32) (helmet, chestplate, leggings, boots item.Stack, ok bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		ok = false
		return
	}
	armour := p.Armour()
	return armour.Helmet(), armour.Chestplate(), armour.Leggings(), armour.Boots(), true
}

func (w *Playback) DespawnPlayer(tx *world.Tx, id uint32) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	tx.RemoveEntity(p)

	p2, _ := w.players[id]
	p2.l.Close(tx)
	delete(w.players, id)
}

func (w *Playback) UpdatePlayerHeldItems(tx *world.Tx, id uint32, mainHand item.Stack, offHand item.Stack) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.SetHeldItems(mainHand, offHand)
}

func (w *Playback) UpdatePlayerArmor(tx *world.Tx, id uint32, helmet, chestplate, leggings, boots item.Stack) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.Armour().Set(helmet, chestplate, leggings, boots)
}

func (w *Playback) DoPlayerSwingArm(tx *world.Tx, id uint32) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.SwingArm()
}

func (w *Playback) SetPlayerSneaking(tx *world.Tx, id uint32, sneaking bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	if sneaking {
		p.StartSneaking()
	} else {
		p.StopSneaking()
	}
}

func (w *Playback) DoPlayerHurt(tx *world.Tx, id uint32) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	for _, v := range player_viewers(p.Player) {
		v.ViewEntityAction(p, entity.HurtAction{})
	}
}

func (w *Playback) DoPlayerEating(tx *world.Tx, id uint32) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	for _, v := range player_viewers(p.Player) {
		v.ViewEntityAction(p, entity.EatAction{})
	}
}

func (w *Playback) SetPlayerUsingItem(tx *world.Tx, id uint32, usingItem bool) {
	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.SetUseItem(usingItem)
}

func (w *Playback) AddParticle(tx *world.Tx, pos mgl64.Vec3, p world.Particle) {
	tx.AddParticle(pos, p)
}

func (w *Playback) PlaySound(tx *world.Tx, pos mgl64.Vec3, s world.Sound) {
	tx.PlaySound(pos, s)
}

func (w *Playback) UpdatePlayerSkin(tx *world.Tx, id uint32, skin skin.Skin) {
	w.skins[id] = skin

	p, ok := w.openPlayer(tx, id)
	if !ok {
		return
	}
	p.SetSkin(skin)
}

// Player returns a player by its ID. If the player does not exist,
// the second return value will be false.
func (w *Playback) Player(id uint32) (*Player, bool) {
	p, ok := w.players[id]
	return p, ok
}

// Entity returns an entity by its ID. If the entity does not exist,
// the second return value will be false.
func (w *Playback) Entity(id uint32) (*Entity, bool) {
	e, ok := w.entities[id]
	return e, ok
}

// openPlayer ...
func (w *Playback) openPlayer(tx *world.Tx, id uint32) (*replayPlayer, bool) {
	h, ok := w.Player(id)
	if !ok {
		return nil, false
	}
	e, ok := h.h.Entity(tx)
	if !ok {
		return nil, false
	}
	p, ok := e.(*replayPlayer)
	return p, ok
}

// openEntity ...
func (w *Playback) openEntity(tx *world.Tx, id uint32) (*entity.Ent, bool) {
	h, ok := w.Entity(id)
	if !ok {
		return nil, false
	}
	e, ok := h.h.Entity(tx)
	if !ok {
		return nil, false
	}
	return e.(*entity.Ent), true
}

// Tick advances the playback by one tick, either forward or backward depending on the reverse setting.
func (w *Playback) Tick(tx *world.Tx) {
	// Check if we should stop processing
	select {
	case <-w.closing:
		return
	default:
	}

	// Don't process if paused
	if w.paused {
		return
	}

	// Handle forward playback
	if !w.reverse {
		// Check if we've reached the end
		if w.playbackTick+1 > w.data.totalTicks {
			w.ended = true
			return
		}

		// Play the next tick and update counter
		w.playTick(tx, w.playbackTick+1)
		w.playbackTick++
		return
	}

	// Handle reverse playback
	// Check if we've reached the beginning
	if w.playbackTick-1 < 0 {
		return
	}

	// Play the previous tick and update counter
	w.reverseTick(tx, w.playbackTick-1)
	w.playbackTick--
}

// playTick executes all actions for the specified tick and stores reverse handlers.
func (w *Playback) playTick(tx *world.Tx, tick uint) {
	actions, ok := w.data.actions[uint32(tick)]
	if !ok {
		return
	}
	reverseHandlers := make([]func(ctx *action.PlayContext), 0, len(actions))
	for _, a := range actions {
		playCtx := action.NewPlayContext(tx, w)
		a.Play(playCtx)
		reverseHandler, hasReverseHandler := playCtx.ReverseHandler()
		if hasReverseHandler {
			reverseHandlers = append(reverseHandlers, reverseHandler)
		}
	}
	w.reverseHandlers[uint32(tick-1)] = reverseHandlers
}

// reverseTick executes the reverse handlers for the specified tick.
func (w *Playback) reverseTick(tx *world.Tx, tick uint) {
	reverseHandlers, ok := w.reverseHandlers[uint32(tick)]
	if !ok {
		return
	}

	for _, h := range reverseHandlers {
		playCtx := action.NewPlayContext(tx, w)
		h(playCtx)
	}

	delete(w.reverseHandlers, uint32(tick))
}

// Close stops the playback and releases resources.
func (w *Playback) Close() {
	w.once.Do(w.doClose)
}

// doClose handles the actual closing logic.
func (w *Playback) doClose() {
	close(w.closing)
	w.running.Wait()
	w.closed.Store(true)
}

// Reversed returns true if the playback is in reverse.
func (w *Playback) Reversed() bool {
	return w.reverse
}

// SetReverse sets the playback to reverse. If reverse is true, the playback
// will play backwards. If false, the playback will play forwards.
func (w *Playback) SetReverse(reverse bool) {
	w.reverse = reverse
}

// Pause pauses the playback.
func (w *Playback) Pause() {
	w.paused = true
}

// Resume resumes the playback.
func (w *Playback) Resume() {
	w.paused = false
}

// Paused returns true if the playback is paused.
func (w *Playback) Paused() bool {
	return w.paused
}

// Speed returns the current playback speed.
func (w *Playback) Speed() float64 {
	return w.speed
}

// SetSpeed sets the playback speed. A value of 1.0 is normal speed,
// values greater than 1.0 are faster, and values less than 1.0 are slower.
// The speed must be greater than 0.
func (w *Playback) SetSpeed(speed float64) {
	if speed <= 0 {
		panic("playback speed must be greater than 0")
	}
	w.speed = speed
}

// FastForward moves the playback forward by the given number of ticks.
func (w *Playback) FastForward(tx *world.Tx, ticks uint) {
	untilTick := min(w.playbackTick+ticks, w.data.totalTicks)

	for i := w.playbackTick + 1; i < untilTick; i++ {
		w.playTick(tx, i)
		w.playbackTick++
	}
}

// Rewind moves the playback backward by the given number of ticks.
func (w *Playback) Rewind(tx *world.Tx, ticks uint) {
	untilTick := max(w.playbackTick-ticks, 0)

	for i := w.playbackTick - 1; i > untilTick; i-- {
		w.reverseTick(tx, i)
		w.playbackTick--
	}
}

// PlaybackTick returns the current tick of the playback.
func (w *Playback) PlaybackTick() uint {
	return w.playbackTick
}

// MaxPlaybackTick returns the maximum tick of the playback.
func (w *Playback) MaxPlaybackTick() uint {
	return w.data.totalTicks - 1
}

// Duration returns the duration of the playback.
func (w *Playback) Duration() time.Duration {
	return time.Duration(w.data.totalTicks) * time.Second / 20
}

// Players ...
func (w *Playback) Players() iter.Seq[*Player] {
	return func(yield func(*Player) bool) {
		for _, p := range w.players {
			if !yield(p) {
				break
			}
		}
	}
}
