package internal

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

func ToSoundID(s world.Sound) (uint32, bool) {
	switch s.(type) {
	case sound.BowShoot:
		return 1, true
	case sound.CrossbowShoot:
		return 2, true
	case sound.ArrowHit:
		return 3, true
	case sound.Teleport:
		return 4, true
	case sound.FireCharge:
		return 5, true
	case sound.Totem:
		return 6, true
	case sound.ItemThrow:
		return 7, true
	}
	return 0, false
}

func FromSoundID(id uint32) (world.Sound, bool) {
	switch id {
	case 1:
		return sound.BowShoot{}, true
	case 2:
		return sound.CrossbowShoot{}, true
	case 3:
		return sound.ArrowHit{}, true
	case 4:
		return sound.Teleport{}, true
	case 5:
		return sound.FireCharge{}, true
	case 6:
		return sound.Totem{}, true
	case 7:
		return sound.ItemThrow{}, true
	default:
		return nil, false
	}
}
