package internal

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

var (
	hashToBlockMapping map[uint64]world.Block
)

func ConstructBlockHashMappings() {
	constructHashToBlockMapping()
}

func constructHashToBlockMapping() {
	// TODO: we don't have world.Blocks() function
}

func HashToBlock(hash uint64) world.Block {
	b, ok := hashToBlockMapping[hash]
	if !ok {
		return block.Air{}
	}
	return b
}
