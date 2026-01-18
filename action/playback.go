package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"time"
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
	SetBlock(tx *world.Tx, pos cube.Pos, b world.Block)
	SpawnPlayer(tx *world.Tx, username, nameTag string, id uint32, pos mgl64.Vec3, rot cube.Rotation, armour [4]item.Stack, heldItems [2]item.Stack)
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
	PlayerNameTag(tx *world.Tx, id uint32) string
	SpawnEntity(tx *world.Tx, id uint32, identifier, nameTag string, pos mgl64.Vec3, rot cube.Rotation, extraData map[string]interface{})
	DespawnEntity(tx *world.Tx, id uint32)
	MoveEntity(tx *world.Tx, id uint32, pos mgl64.Vec3, rot cube.Rotation)
	EntityPosition(tx *world.Tx, id uint32) (mgl64.Vec3, bool)
	EntityRotation(tx *world.Tx, id uint32) (cube.Rotation, bool)
	EntityIdentifier(id uint32) (string, bool)
	EntityNameTag(tx *world.Tx, id uint32) string
	EntityExtraData(id uint32) (map[string]any, bool)
	SetPlayerNameTag(tx *world.Tx, id uint32, nameTag string)
	SetEntityNameTag(tx *world.Tx, id uint32, nameTag string)
	Liquid(tx *world.Tx, pos cube.Pos) (world.Liquid, bool)
	SetLiquid(tx *world.Tx, pos cube.Pos, l world.Liquid)
	UpdateChestState(tx *world.Tx, pos cube.Pos, open bool)
	ChestState(tx *world.Tx, pos cube.Pos) bool
	StartCrackBlock(tx *world.Tx, pos cube.Pos, duration time.Duration)
	StopCrackBlock(tx *world.Tx, pos cube.Pos)
	ContinueCrackBlock(tx *world.Tx, pos cube.Pos, duration time.Duration)
	Emote(tx *world.Tx, id uint32, emoteId uuid.UUID)
	PlayerVisible(tx *world.Tx, id uint32) bool
	SetPlayerVisibility(tx *world.Tx, id uint32, visible bool)
	PlayerSprinting(tx *world.Tx, id uint32) bool
	SetPlayerSprinting(tx *world.Tx, id uint32, sprinting bool)
	PlayerGliding(tx *world.Tx, id uint32) bool
	SetPlayerGliding(tx *world.Tx, id uint32, gliding bool)
	PlayerSwimming(tx *world.Tx, id uint32) bool
	SetPlayerSwimming(tx *world.Tx, id uint32, swimming bool)
	PlayerCrawling(tx *world.Tx, id uint32) bool
	SetPlayerCrawling(tx *world.Tx, id uint32, crawling bool)
	DoPlayerTotemUse(tx *world.Tx, id uint32)
	PlayerOnFire(tx *world.Tx, id uint32) bool
	SetPlayerOnFire(tx *world.Tx, id uint32, onFire bool)
	PlayerVisibleEffects(tx *world.Tx, id uint32) ([]int, bool)
	SetPlayerVisibleEffects(tx *world.Tx, id uint32, effectIDs []int)
	DoPlayerCriticalHit(tx *world.Tx, id uint32)
	DoPlayerEnchantedHit(tx *world.Tx, id uint32)
	DoFireworkExplosion(tx *world.Tx, id uint32)
	DoArrowShake(tx *world.Tx, id uint32)
}
