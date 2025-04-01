package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Playback interface {
	Block(tx *world.Tx, pos cube.Pos) world.Block
	PlayerPosition(tx *world.Tx, id uint32) (mgl64.Vec3, bool)
	PlayerRotation(tx *world.Tx, id uint32) (cube.Rotation, bool)
	PlayerArmours(tx *world.Tx, id uint32) (helmet, chestplate, leggings, boots item.Stack, ok bool)
	PlayerHeldItems(tx *world.Tx, id uint32) (mainHand, offHand item.Stack, ok bool)
	PlayerSneaking(tx *world.Tx, id uint32) bool
	PlayerUsingItem(tx *world.Tx, id uint32) bool
	PlayerSkin(id uint32) (skin.Skin, bool)
	MovePlayer(tx *world.Tx, id uint32, pos mgl64.Vec3, rot cube.Rotation)
	SetBlock(tx *world.Tx, pos cube.Pos, b world.Block, layer uint8)
	SpawnPlayer(tx *world.Tx, name string, id uint32, pos mgl64.Vec3, rot cube.Rotation, armour [4]item.Stack, heldItems [2]item.Stack)
	DespawnPlayer(tx *world.Tx, id uint32)
	UpdatePlayerHeldItems(tx *world.Tx, id uint32, mainHand item.Stack, offHand item.Stack)
	UpdatePlayerArmours(tx *world.Tx, id uint32, helmet, chestplate, leggings, boots item.Stack)
	DoPlayerSwingArm(tx *world.Tx, id uint32)
	SetPlayerSneaking(tx *world.Tx, id uint32, sneaking bool)
	DoPlayerHurt(tx *world.Tx, id uint32)
	DoPlayerEating(tx *world.Tx, id uint32)
	SetPlayerUsingItem(tx *world.Tx, id uint32, usingItem bool)
	AddParticle(tx *world.Tx, pos mgl64.Vec3, p world.Particle)
	PlaySound(tx *world.Tx, pos mgl64.Vec3, s world.Sound)
	UpdatePlayerSkin(tx *world.Tx, id uint32, skin skin.Skin)
	PlayerName(id uint32) string
}
