package internal

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

func ToParticleID(p world.Particle) (uint32, bool) {
	switch p.(type) {
	case particle.HugeExplosion:
		return 1, true
	case particle.BoneMeal:
		return 2, true
	}
	// TODO: implement more particles
	return 0, false
}

func FromParticleID(id uint32) (world.Particle, bool) {
	switch id {
	case 1:
		return particle.HugeExplosion{}, true
	case 2:
		return particle.BoneMeal{}, true
	}
	return nil, false
}
