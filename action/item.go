package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Item struct {
	Hash uint32
}

func ItemFromStack(stack item.Stack) Item {
	return Item{Hash: internal.ItemToHash(stack.Item())}
}

func (i Item) Marshal(io protocol.IO) {
	io.Uint32(&i.Hash)
}

func (i Item) ToStack() item.Stack {
	it := internal.HashToItem(i.Hash)
	return item.NewStack(it, 1)
}
