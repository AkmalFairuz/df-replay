package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Block struct {
	Hash   uint32
	HasNBT bool
	NBT    map[string]any
}

func (b *Block) Marshal(io protocol.IO) {
	io.Uint32(&b.Hash)
	io.Bool(&b.HasNBT)
	if b.HasNBT {
		io.NBT(&b.NBT, nbt.NetworkLittleEndian)
	}
}

func FromBlock(block world.Block) Block {
	ret := Block{
		Hash: internal.BlockToHash(block),
	}

	if v, ok := block.(world.NBTer); ok {
		ret.HasNBT = true
		ret.NBT = v.EncodeNBT()
	} else {
		ret.HasNBT = false
	}

	return ret
}

func (b *Block) ToBlock() world.Block {
	ret := internal.HashToBlock(b.Hash)
	if b.HasNBT {
		if v, ok := ret.(world.NBTer); ok {
			ret = v.DecodeNBT(b.NBT).(world.Block)
		}
	}
	return ret
}

func blockPosToCubePos(pos protocol.BlockPos) cube.Pos {
	return cube.Pos{int(pos[0]), int(pos[1]), int(pos[2])}
}
