package internal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1"
)

var (
	hashToItemMapping map[uint32]world.Item
	itemToHashMapping map[string]uint32
)

func ConstructItemHashMappings() {
	hashToItemMapping = make(map[uint32]world.Item)
	itemToHashMapping = make(map[string]uint32)

	for _, it := range world.Items() {
		name, meta := it.EncodeItem()
		itemStr := fmt.Sprintf("%s:%d", name, meta)
		hash := fnv1.HashString32(itemStr)
		hashToItemMapping[hash] = it
		itemToHashMapping[itemStr] = hash
	}
}

func ItemToString(it world.Item) string {
	name, meta := it.EncodeItem()
	return fmt.Sprintf("%s:%d", name, meta)
}

func ItemToHash(it world.Item) uint32 {
	if it == nil {
		it = block.Air{}
	}
	hash, ok := itemToHashMapping[ItemToString(it)]
	if !ok {
		// slow hash
		return fnv1.HashString32(ItemToString(it))
	}
	return hash
}

func HashToItem(hash uint32) world.Item {
	it, ok := hashToItemMapping[hash]
	if !ok {
		return block.Air{}
	}
	return it
}
