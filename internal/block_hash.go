package internal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1"
	"sort"
	"strings"
)

var (
	hashToBlockMapping map[uint32]world.Block
	blockToHashMapping map[uint32]uint32
)

func ConstructBlockHashMappings() {
	blocks := world.Blocks()
	hashToBlockMapping = make(map[uint32]world.Block, len(blocks))
	blockToHashMapping = make(map[uint32]uint32, len(blocks))
	for _, b := range blocks {
		hash := computeHash(b)
		hashToBlockMapping[hash] = b
		blockToHashMapping[world.BlockRuntimeID(b)] = hash
	}
}

func computeHash(b world.Block) uint32 {
	name, properties := b.EncodeBlock()
	type property struct {
		key string
		val any
	}
	var props []property
	for k, v := range properties {
		props = append(props, property{k, v})
	}
	sort.Slice(props, func(i, j int) bool {
		return props[i].key < props[j].key
	})
	toHash := strings.Builder{}
	toHash.WriteString(name)
	toHash.WriteString("|")
	for _, p := range props {
		toHash.WriteString(p.key)
		toHash.WriteString("=")
		toHash.WriteString(fmt.Sprintf("%v", p.val))
		toHash.WriteString("|")
	}
	return fnv1.HashString32(toHash.String())
}

func HashToBlock(hash uint32) world.Block {
	b, ok := hashToBlockMapping[hash]
	if !ok {
		return block.Air{}
	}
	return b
}

func BlockToHash(b world.Block) uint32 {
	k := world.BlockRuntimeID(b)
	hash, ok := blockToHashMapping[k]
	if !ok {
		return 0
	}
	return hash
}
