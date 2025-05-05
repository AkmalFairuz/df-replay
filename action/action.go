package action

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Action interface {
	ID() uint8
	Marshal(io protocol.IO)
	Play(ctx *PlayContext)
}

var (
	actionPool = map[uint8]func() Action{
		IDPlayerSpawn:         func() Action { return &PlayerSpawn{} },
		IDPlayerMove:          func() Action { return &PlayerMove{} },
		IDPlayerAnimate:       func() Action { return &PlayerAnimate{} },
		IDPlayerDespawn:       func() Action { return &PlayerDespawn{} },
		IDSetBlock:            func() Action { return &SetBlock{} },
		IDPlayerHandChange:    func() Action { return &PlayerHandChange{} },
		IDPlayerArmorChange:   func() Action { return &PlayerArmorChange{} },
		IDBreakBlock:          func() Action { return &BreakBlock{} },
		IDPlaceBlock:          func() Action { return &PlaceBlock{} },
		IDPlayerSkin:          func() Action { return &PlayerSkin{} },
		IDEntitySpawn:         func() Action { return &EntitySpawn{} },
		IDEntityDespawn:       func() Action { return &EntityDespawn{} },
		IDEntityMove:          func() Action { return &EntityMove{} },
		IDPlayerNameTagUpdate: func() Action { return &PlayerNameTagUpdate{} },
		IDEntityNameTagUpdate: func() Action { return &EntityNameTagUpdate{} },
		IDSetLiquid:           func() Action { return &SetLiquid{} },
		IDChestUpdate:         func() Action { return &ChestUpdate{} },
	}
)

// Read ...
func Read(io *protocol.Reader, act *Action) (err error) {
	var id uint8
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("action read id=%d panic: %v", id, r)
		}
	}()
	io.Uint8(&id)
	if f, ok := actionPool[id]; ok {
		*act = f()
		(*act).Marshal(io)
		return nil
	}
	return fmt.Errorf("unknown action id: %d", id)
}

// Write ...
func Write(io *protocol.Writer, act Action) {
	io.Uint8(lo.ToPtr(act.ID()))
	act.Marshal(io)
}
