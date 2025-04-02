package replay

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server/item"
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
		data := getEntityHandleData(h, "Data").(*replayEntityData)
		if data.Identifier == "minecraft:item" {
			ctx.Cancel()
			stack := item.NewStack(internal.HashToItem(uint32(data.ExtraData["Item"].(int32))), int(data.ExtraData["ItemCount"].(int32)))
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
		pk.EntityType = data.Identifier
	}
}
