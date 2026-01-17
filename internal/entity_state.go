package internal

import "github.com/df-mc/dragonfly/server/world"

type EntityState struct {
	NameTag string
}

func GetEntityState(e world.Entity) (s EntityState) {
	if v, ok := e.(interface{ NameTag() string }); ok {
		s.NameTag = v.NameTag()
	}
	return
}
