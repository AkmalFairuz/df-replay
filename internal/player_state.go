package internal

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
)

type PlayerState struct {
	Invisible bool
	UsingItem bool
	Sneaking  bool
	Swimming  bool
	Crawling  bool
	Gliding   bool
	Sprinting bool
	NameTag   string
	OnFire    bool

	VisibleParticleEffectIDs []int
}

func GetPlayerState(p *player.Player) (s PlayerState) {
	s.Invisible = p.Invisible()
	s.UsingItem = p.UsingItem()
	s.Sneaking = p.Sneaking()
	s.Swimming = p.Swimming()
	s.Crawling = p.Crawling()
	s.Gliding = p.Gliding()
	s.Sprinting = p.Sprinting()
	s.NameTag = p.NameTag()
	s.OnFire = p.OnFireDuration() > 0
	for _, e := range p.Effects() {
		if e.ParticlesHidden() {
			continue
		}
		effectId, ok := effect.ID(e.Type())
		if !ok {
			continue
		}
		s.VisibleParticleEffectIDs = append(s.VisibleParticleEffectIDs, effectId)
	}
	return
}

func EqualEffectIDs(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[int]int)
	for _, id := range a {
		m[id]++
	}
	for _, id := range b {
		m[id]--
		if m[id] < 0 {
			return false
		}
	}
	return true
}
