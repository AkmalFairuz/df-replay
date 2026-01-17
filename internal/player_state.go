package internal

import (
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
	return
}
