package action

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Item struct {
	Hash       uint32
	HasEnchant bool
}

func ItemFromStack(stack item.Stack) Item {
	return Item{
		Hash:       internal.ItemToHash(stack.Item()),
		HasEnchant: len(stack.Enchantments()) > 0,
	}
}

func (i *Item) Marshal(io protocol.IO) {
	io.Uint32(&i.Hash)
	io.Bool(&i.HasEnchant)
}

type nopEnchantment struct{}

func (n nopEnchantment) Name() string                                        { return "nop" }
func (n nopEnchantment) MaxLevel() int                                       { return 1 }
func (n nopEnchantment) Cost(int) (int, int)                                 { return 0, 0 }
func (n nopEnchantment) Rarity() item.EnchantmentRarity                      { return item.EnchantmentRarityCommon }
func (n nopEnchantment) CompatibleWithEnchantment(item.EnchantmentType) bool { return true }
func (n nopEnchantment) CompatibleWithItem(world.Item) bool                  { return true }

func (i *Item) ToStack() item.Stack {
	it := internal.HashToItem(i.Hash)
	ret := item.NewStack(it, 1)
	if i.HasEnchant {
		return ret.WithEnchantments(item.NewEnchantment(nopEnchantment{}, 1))
	}
	return ret
}
