package replay

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"reflect"
	"unsafe"
	_ "unsafe"
)

// updatePlayerData updates the world.EntityData of a player.
func updatePlayerEntityData(p *player.Player, field string, val any) {
	rf := reflect.ValueOf(p).Elem().FieldByName("data")
	f := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Elem().FieldByName(field)
	f.Set(reflect.ValueOf(val))
}

// updatePlayerData updates the player.playerData field of the player.Player struct.
func updatePlayerData(p *player.Player, field string, val any) {
	rf := reflect.ValueOf(p).Elem().FieldByName("playerData")
	f := reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Elem().FieldByName(field)
	f.Set(reflect.ValueOf(val))
}

//go:linkname player_viewers github.com/df-mc/dragonfly/player.(*Player).viewers
func player_viewers(*player.Player) []world.Viewer

//go:linkname player_updateState github.com/df-mc/dragonfly/player.(*Player).updateState
func player_updateState(*player.Player)

//go:linkname player_canRelease github.com/df-mc/dragonfly/player.(*Player).canRelease
func player_canRelease(*player.Player) bool
