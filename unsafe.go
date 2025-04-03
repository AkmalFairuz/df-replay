package replay

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"reflect"
	"unsafe"
)

// updatePlayerData updates the world.EntityData of a player.
func updatePlayerEntityData(p *player.Player, field string, val any) {
	rf := reflect.ValueOf(p).Elem().FieldByName("data")
	f := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Elem().FieldByName(field)
	if !f.CanSet() {
		reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(val))
		return
	}
	f.Set(reflect.ValueOf(val))
}

// updateEntEntityData updates the world.EntityData of an entity.
func updateEntEntityData(e *entity.Ent, field string, val any) {
	rf := reflect.ValueOf(e).Elem().FieldByName("data")
	f := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Elem().FieldByName(field)
	if !f.CanSet() {
		reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(val))
		return
	}
	f.Set(reflect.ValueOf(val))
}

// updatePlayerData updates the player.playerData field of the player.Player struct.
func updatePlayerData(p *player.Player, field string, val any) {
	rf := reflect.ValueOf(p).Elem().FieldByName("playerData")
	f := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Elem().FieldByName(field)
	if !f.CanSet() {
		reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(val))
		return
	}
	f.Set(reflect.ValueOf(val))
}

func toAny(a any) any {
	return a
}

// getSessionByHandle returns the session of a player by its entity handle.
func getSessionByHandle(h *world.EntityHandle) *session.Session {
	// detect venity fork
	if v, ok := toAny(h).(interface{ EntityData() world.EntityData }); ok {
		if v2, ok := v.EntityData().Data.(interface{ Session() *session.Session }); ok {
			return v2.Session()
		}
	}
	rf := reflect.ValueOf(h).Elem().FieldByName("data")
	rs := rf.FieldByName("Data").Elem().Elem().FieldByName("s")
	return reflect.NewAt(rs.Type(), unsafe.Pointer(rs.UnsafeAddr())).Elem().Interface().(*session.Session)
}

// getEntityHandleData ...
func getEntityHandleData(h *world.EntityHandle, field string) any {
	rf := reflect.ValueOf(h).Elem().FieldByName("data")
	f := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().FieldByName(field)
	if !f.CanSet() {
		return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Interface()
	}
	return f.Interface()
}

//go:linkname player_viewers github.com/df-mc/dragonfly/server/player.(*Player).viewers
func player_viewers(*player.Player) []world.Viewer

//go:linkname player_updateState github.com/df-mc/dragonfly/server/player.(*Player).updateState
func player_updateState(*player.Player)

//go:linkname player_canRelease github.com/df-mc/dragonfly/server/player.(*Player).canRelease
func player_canRelease(*player.Player) bool

//go:linkname session_writePacket github.com/df-mc/dragonfly/server/session.(*Session).writePacket
func session_writePacket(*session.Session, packet.Packet)

//go:linkname session_entityFromRuntimeID github.com/df-mc/dragonfly/server/session.(*Session).entityFromRuntimeID
func session_entityFromRuntimeID(*session.Session, uint64) (*world.EntityHandle, bool)

//go:linkname instanceFromItem github.com/df-mc/dragonfly/server/session.instanceFromItem
func instanceFromItem(item.Stack) protocol.ItemInstance
