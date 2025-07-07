package action

import (
	"bytes"
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ItemFlagHasEnchant uint8 = 1 << iota
	ItemFlagHasNBT
)

type Item struct {
	Hash  uint32
	Flags uint8
	NBT   []byte
}

func ItemFromStack(stack item.Stack) Item {
	nbtBytes := bytes.NewBuffer(nil)
	enc := nbt.NewEncoderWithEncoding(nbtBytes, nbt.NetworkLittleEndian) // use network encoding to save space
	if nbter, ok := stack.Item().(world.NBTer); ok {
		_ = enc.Encode(nbter.EncodeNBT())
	}
	flags := uint8(0)
	if len(stack.Enchantments()) > 0 {
		flags |= ItemFlagHasEnchant
	}
	if nbtBytes.Len() > 0 {
		flags |= ItemFlagHasNBT
	}
	return Item{
		Hash:  internal.ItemToHash(stack.Item()),
		Flags: flags,
		NBT:   nbtBytes.Bytes(),
	}
}

func (i *Item) Marshal(io protocol.IO) {
	io.Uint32(&i.Hash)
	io.Uint8(&i.Flags)
	if i.Flags&ItemFlagHasNBT != 0 {
		io.ByteSlice(&i.NBT)
	}
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
	if i.Flags&ItemFlagHasNBT != 0 {
		dec := nbt.NewDecoderWithEncoding(bytes.NewReader(i.NBT), nbt.NetworkLittleEndian)
		var nbtData map[string]any
		if nbter, ok := it.(world.NBTer); ok {
			if err := dec.Decode(&nbtData); err == nil {
				it = nbter.DecodeNBT(nbtData).(world.Item)
			}
		}
	}
	ret := item.NewStack(it, 1)
	if i.Flags&ItemFlagHasNBT != 0 {
		return ret.WithEnchantments(item.NewEnchantment(nopEnchantment{}, 1))
	}
	return ret
}
