package internal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1a"
	"sort"
	"strings"
)

var (
	hashToBlockMapping map[uint32]world.Block
	blockToHashMapping map[uint64]uint32
)

func ConstructBlockHashMappings() {
	for _, b := range world.Blocks() {
		hash := computeHash(b)
		hashToBlockMapping[hash] = b
		blockToHashMapping[world.BlockHash(b)] = hash
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
	return fnv1a.HashString32(toHash.String())
}

func HashToBlock(hash uint32) world.Block {
	b, ok := hashToBlockMapping[hash]
	if !ok {
		return block.Air{}
	}
	return b
}

func BlockToHash(b world.Block) uint32 {
	k := world.BlockHash(b)
	hash, ok := blockToHashMapping[k]
	if !ok {
		return 0
	}
	return hash
}
