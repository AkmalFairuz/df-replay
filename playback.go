package replay

import (
	"github.com/akmalfairuz/df-replay/action"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync/atomic"
	"time"
)

// Playback ...
type Playback struct {
	w    *world.World
	data *Data

	totalTicks   uint
	playbackTick uint
	stopped      bool
	paused       bool
	speed        int
	closed       atomic.Bool
	reverse      bool
	ended        bool

	players         map[uint32]*Player
	entities        map[uint32]*Entity
	skins           map[uint32]skin.Skin
	reverseHandlers map[uint32][]func(ctx *action.PlayContext)
}

// Compile time check to ensure that Playback implements action.Playback.
var _ action.Playback = (*Playback)(nil)

// NewPlayback ...
func NewPlayback(w *world.World, data *Data) *Playback {
	return &Playback{
		w:               w,
		data:            data,
		players:         make(map[uint32]*Player),
		entities:        make(map[uint32]*Entity),
		skins:           make(map[uint32]skin.Skin),
		reverseHandlers: make(map[uint32][]func(ctx *action.PlayContext)),
	}
}

// Play ...
func (w *Playback) Play() {
	go func() {
		ticker := time.NewTicker(time.Second / 20)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				<-w.w.Exec(w.Tick)
			}
		}
	}()
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
}

func (w *Playback) SetBlock(tx *world.Tx, pos cube.Pos, b world.Block, _ uint8) {
	tx.SetBlock(pos, b, &world.SetOpts{
		DisableBlockUpdates:       true,
		DisableLiquidDisplacement: true,
	})
}

func (w *Playback) SpawnPlayer(tx *world.Tx, name string, id uint32, pos mgl64.Vec3, rot cube.Rotation, armour [4]item.Stack, heldItems [2]item.Stack) {
	opts := &world.EntitySpawnOpts{
		Position: pos,
		Rotation: rot,
		NameTag:  name,
	}
	conf := player.Config{
		Name:      name,
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
	w.players[id] = &Player{
		id:   id,
		name: name,
		h:    h,
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
	if usingItem {
		p.UseItem()
	}
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

// Tick ...
func (w *Playback) Tick(tx *world.Tx) {
	if w.paused {
		return
	}

	if !w.reverse {
		if w.playbackTick+1 >= w.totalTicks {
			w.ended = true
			return
		}
		actions, ok := w.data.actions[uint32(w.playbackTick)]
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
		w.reverseHandlers[uint32(w.playbackTick)] = reverseHandlers
		w.playbackTick++
		return
	}

	if w.playbackTick-1 < 0 {
		return
	}

	reverseHandlers, ok := w.reverseHandlers[uint32(w.playbackTick)]
	if !ok {
		return
	}
	for _, h := range reverseHandlers {
		playCtx := action.NewPlayContext(tx, w)
		h(playCtx)
	}
	w.playbackTick--
}
