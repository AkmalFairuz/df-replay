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
		if data.identifier == "minecraft:item" {
			ctx.Cancel()
			stack := item.NewStack(internal.HashToItem(uint32(data.extraData["Item"].(int64))), int(data.extraData["ItemCount"].(int32)))
			session_writePacket(s, &packet.AddItemActor{
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				Item:            instanceFromItem(stack),
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				EntityMetadata:  pk.EntityMetadata,
			})
			return
		}
		if _, isTextType := data.extraData["IsTextType"]; isTextType {
			pk.EntityMetadata[protocol.EntityDataKeyVariant] = world.BlockRuntimeID(block.Air{})
		}
		pk.EntityType = data.identifier
	}
}
