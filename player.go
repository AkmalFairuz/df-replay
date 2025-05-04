package replay

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Player struct {
	id   uint32
	name string
	h    *world.EntityHandle
	l    *world.Loader
}

func (p *Player) ExecEntity(tx *world.Tx, fn func(*world.Tx, world.Entity)) bool {
	e, ok := p.h.Entity(tx)
	if !ok {
		return false
	}
	if e, ok := e.(*replayPlayer); ok {
		fn(tx, e)
		return true
	}
	return p.h.ExecWorld(fn)
}

var playerType = ptype{t: player.Type}

func init() {}

type ptype struct {
	t world.EntityType
}

func (p ptype) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	ret := p.t.Open(tx, handle, data).(*player.Player)
	return &replayPlayer{Player: ret}
}

func (p ptype) EncodeEntity() string {
	return "replay_player"
}

func (p ptype) NetworkEncodeEntity() string {
	return "minecraft:player" // just in case this function is called, even it should not be called
}

func (p ptype) NetworkOffset() float64 { return 1.621 }

func (p ptype) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3) // TODO: use the real player bbox
}

func (p ptype) DecodeNBT(m map[string]any, data *world.EntityData) {
	p.t.DecodeNBT(m, data)
}

func (p ptype) EncodeNBT(data *world.EntityData) map[string]any {
	return p.t.EncodeNBT(data)
}

type replayPlayer struct {
	*player.Player

	useItem bool
}

func (p *replayPlayer) Tick(*world.Tx, int64) {}

func (p *replayPlayer) Hurt(float64, world.DamageSource) (float64, bool) { return 0, false }

func (p *replayPlayer) SetUseItem(useItem bool) {
	p.useItem = useItem
	player_updateState(p.Player)
}

func (p *replayPlayer) UsingItem() bool {
	return p.useItem
}

func (p *replayPlayer) MoveSmooth(pos mgl64.Vec3, rot cube.Rotation) {
	// detect venity fork
	if v, ok := toAny(p.Player).(interface {
		SetPosAndRotNoUpdate(pos mgl64.Vec3, rot cube.Rotation)
	}); ok {
		v.SetPosAndRotNoUpdate(pos, rot)
	} else {
		updatePlayerEntityData(p.Player, "Pos", pos)
		updatePlayerEntityData(p.Player, "Rot", rot)
	}

	for _, v := range player_viewers(p.Player) {
		v.ViewEntityMovement(p.Player, pos, rot, false)
	}
}
