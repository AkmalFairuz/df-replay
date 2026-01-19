package replay

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetHandler struct{}

func (packetHandler) HandleClientPacket(ctx *intercept.Context, pk packet.Packet) {}

func (packetHandler) HandleServerPacket(ctx *intercept.Context, pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.AddActor:
		if pk.EntityType != "replay_entity" {
			return
		}
		s := getSessionByHandle(ctx.Val())
		h, ok := session_entityFromRuntimeID(s, pk.EntityRuntimeID)
		if !ok {
			ctx.Cancel()
			return
		}
		var data *entityBehaviour
		// detect venity fork
		if v, ok := toAny(h).(interface{ EntityData() world.EntityData }); ok {
			data = v.EntityData().Data.(*entityBehaviour)
		} else {
			data = getEntityHandleData(h, "Data").(*entityBehaviour)
		}
		switch data.identifier {
		case "minecraft:item":
			ctx.Cancel()
			stack := item.NewStack(internal.HashToItem(uint32(data.extraData["Item"].(int64))), int(data.extraData["ItemCount"].(int32)))
			conn := getConnBySession(s)
			_ = conn.WritePacket(&packet.AddItemActor{
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				Item:            instanceFromItem(stack),
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				EntityMetadata:  pk.EntityMetadata,
			})
			return
		case "minecraft:tnt":
			fuseTime, ok := data.extraData["FuseTime"]
			if !ok {
				pk.EntityMetadata[protocol.EntityDataKeyFuseTime] = int32(80)
			} else {
				pk.EntityMetadata[protocol.EntityDataKeyFuseTime] = fuseTime.(int32) / 50
			}
			protocol.EntityMetadata(pk.EntityMetadata).SetFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagIgnited)
		case "minecraft:falling_block":
			blockHash, ok := data.extraData["Block"]
			if ok {
				pk.EntityMetadata[protocol.EntityDataKeyVariant] = world.BlockRuntimeID(internal.HashToBlock(uint32(blockHash.(int32))))
			}
		case "minecraft:firework":
			f, ok := data.extraData["Item"]
			if ok {
				it := item.Firework{}.DecodeNBT(f.(map[string]any)).(item.Firework)
				pk.EntityMetadata[protocol.EntityDataKeyDisplayTileRuntimeID] = nbtconv_WriteItem(item.NewStack(it, 1), false)
			}
		case "minecraft:splash_potion", "minecraft:arrow":
			potionId, ok := data.extraData["PotionID"]
			if ok {
				pk.EntityMetadata[protocol.EntityDataKeyAuxValueData] = int16(potionId.(int32))
				if tip := uint8(potionId.(int32)); tip > 4 {
					pk.EntityMetadata[protocol.EntityDataKeyCustomDisplay] = tip + 1
				}
			}
		}
		if _, isTextType := data.extraData["IsTextType"]; isTextType {
			pk.EntityMetadata[protocol.EntityDataKeyVariant] = int32(world.BlockRuntimeID(block.Air{}))
		}
		pk.EntityType = data.identifier
	}
}
