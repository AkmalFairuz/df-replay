package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func blockPosToCubePos(pos protocol.BlockPos) cube.Pos {
	return cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}
}
