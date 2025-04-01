package internal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1a"
)

var (
	hashToItemMapping map[uint32]world.Item
	itemToHashMapping map[string]uint32
)

func ConstructItemHashMappings() {
	constructHashToItemMapping()
	constructItemToHashMapping()
}

func constructHashToItemMapping() {
	items := world.Items()
	hashToItemMapping = make(map[uint32]world.Item, len(items))
	for _, it := range items {
		hash := fnv1a.HashString32(ItemToString(it))
		hashToItemMapping[hash] = it
	}
}

func constructItemToHashMapping() {
	items := world.Items()
	itemToHashMapping = make(map[string]uint32, len(items))
	for _, it := range items {
		hash := fnv1a.HashString32(ItemToString(it))
		itemToHashMapping[ItemToString(it)] = hash
	}
}

func ItemToString(it world.Item) string {
	name, meta := it.EncodeItem()
	return fmt.Sprintf("%s:%d", name, meta)
}

func ItemToHash(it world.Item) uint32 {
	hash, ok := itemToHashMapping[ItemToString(it)]
	if !ok {
		// slow hash
		return fnv1a.HashString32(ItemToString(it))
	}
	return hash
}

func HashToItem(hash uint32) world.Item {
	it, ok := hashToItemMapping[hash]
	if !ok {
		return nil
	}
	return it
}
