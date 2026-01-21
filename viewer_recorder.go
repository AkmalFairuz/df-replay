package replay

import (
	"github.com/akmalfairuz/df-replay/internal"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"sync"
)

type RecorderViewer struct {
	world.NopViewer

	r *Recorder

	playerStatesMu sync.Mutex
	playerStates   map[uuid.UUID]internal.PlayerState

	entityStatesMu sync.Mutex
	entityStates   map[uuid.UUID]internal.EntityState
}

func NewRecorderViewer(r *Recorder) world.Viewer {
	return &RecorderViewer{
		r:            r,
		playerStates: make(map[uuid.UUID]internal.PlayerState),
		entityStates: make(map[uuid.UUID]internal.EntityState),
	}
}

func (r *RecorderViewer) ViewEntity(e world.Entity) {
	switch e := e.(type) {
	case *player.Player:
		if !e.GameMode().Visible() {
			return
		}
		r.r.AddPlayer(e)
	default:
		r.r.AddEntity(e)
	}
}

func (r *RecorderViewer) ViewEntityGameMode(e world.Entity) {
	switch e := e.(type) {
	case *player.Player:
		if !e.GameMode().Visible() {
			r.r.RemovePlayer(e)
			r.playerStatesMu.Lock()
			delete(r.playerStates, e.UUID())
			r.playerStatesMu.Unlock()
		} else {
			r.playerStatesMu.Lock()
			_, ok := r.playerStates[e.UUID()]
			r.playerStatesMu.Unlock()
			if !ok {
				r.r.AddPlayer(e)
			}
		}
	}
}

func (r *RecorderViewer) HideEntity(e world.Entity) {
	switch e := e.(type) {
	case *player.Player:
		r.r.RemovePlayer(e)
		r.playerStatesMu.Lock()
		delete(r.playerStates, e.UUID())
		r.playerStatesMu.Unlock()
	default:
		r.r.RemoveEntity(e)
		r.entityStatesMu.Lock()
		delete(r.entityStates, e.H().UUID())
		r.entityStatesMu.Unlock()
	}
}

func (r *RecorderViewer) ViewEntityMovement(e world.Entity, pos mgl64.Vec3, rot cube.Rotation, onGround bool) {
	switch e := e.(type) {
	case *player.Player:
		if !e.GameMode().Visible() { // early return if the player is invisible
			return
		}
		r.r.PushPlayerMovement(e, pos, rot)
	default:
		r.r.PushEntityMovement(e, pos, rot)
	}
}

func (r *RecorderViewer) ViewEntityTeleport(e world.Entity, pos mgl64.Vec3) {
	switch e := e.(type) {
	case *player.Player:
		r.r.PushPlayerMovement(e, pos, e.Rotation())
	default:
		r.r.PushEntityMovement(e, pos, e.Rotation())
	}
}

func (r *RecorderViewer) ViewEntityItems(e world.Entity) {
	switch e := e.(type) {
	case *player.Player:
		mainHand, offHand := e.HeldItems()
		r.r.PushPlayerHandChange(e, mainHand, offHand)
	}
}

func (r *RecorderViewer) ViewEntityArmour(e world.Entity) {
	switch e := e.(type) {
	case *player.Player:
		r.r.PushPlayerArmorChange(e)
	}
}

func (r *RecorderViewer) ViewEntityAction(e world.Entity, a world.EntityAction) {
	switch e := e.(type) {
	case *player.Player:
		switch a.(type) {
		case entity.SwingArmAction:
			r.r.PushPlayerSwingArm(e)
		case entity.HurtAction:
			r.r.PushPlayerHurt(e)
		case entity.EatAction:
			r.r.PushPlayerEating(e)
		case entity.TotemUseAction:
			r.r.PushPlayerTotemUse(e)
		case entity.CriticalHitAction:
			r.r.PushPlayerCriticalHit(e)
		case entity.EnchantedHitAction:
			r.r.PushPlayerEnchantedHit(e)
		}
	default:
		switch a.(type) {
		case entity.FireworkExplosionAction:
			r.r.PushEntityFireworkExplosion(e)
		case entity.ArrowShakeAction:
			r.r.PushEntityArrowShake(e)
		}
	}
}

func (r *RecorderViewer) ViewEntityState(e world.Entity) {
	switch e := e.(type) {
	case *player.Player:
		if !e.GameMode().Visible() { // early return if the player is invisible
			return
		}
		s := internal.GetPlayerState(e)
		r.playerStatesMu.Lock()
		prev, ok := r.playerStates[e.UUID()]
		r.playerStates[e.UUID()] = s
		r.playerStatesMu.Unlock()

		if !ok {
			// TODO: higher disk usage
			r.r.PushPlayerSneaking(e, s.Sneaking)
			r.r.PushPlayerUsingItem(e, s.UsingItem)
			r.r.PushPlayerVisibility(e, !s.Invisible)
			r.r.PushPlayerSprinting(e, s.Sprinting)
			r.r.PushPlayerGliding(e, s.Gliding)
			r.r.PushPlayerCrawling(e, s.Crawling)
			r.r.PushPlayerSwimming(e, s.Swimming)
			r.r.PushSetPlayerNameTag(e, s.NameTag)
			r.r.PushPlayerOnFire(e, s.OnFire)
			r.r.PushPlayerSetVisibleEffects(e, s.VisibleParticleEffectIDs)
			return
		}

		if prev.Sneaking != s.Sneaking {
			r.r.PushPlayerSneaking(e, s.Sneaking)
		}
		if prev.UsingItem != s.UsingItem {
			r.r.PushPlayerUsingItem(e, s.UsingItem)
		}
		if prev.Invisible != s.Invisible {
			r.r.PushPlayerVisibility(e, !s.Invisible)
		}
		if prev.Sprinting != s.Sprinting {
			r.r.PushPlayerSprinting(e, s.Sprinting)
		}
		if prev.Gliding != s.Gliding {
			r.r.PushPlayerGliding(e, s.Gliding)
		}
		if prev.Crawling != s.Crawling {
			r.r.PushPlayerCrawling(e, s.Crawling)
		}
		if prev.Swimming != s.Swimming {
			r.r.PushPlayerSwimming(e, s.Swimming)
		}
		if prev.NameTag != s.NameTag {
			r.r.PushSetPlayerNameTag(e, s.NameTag)
		}
		if prev.OnFire != s.OnFire {
			r.r.PushPlayerOnFire(e, s.OnFire)
		}
		if len(s.VisibleParticleEffectIDs) > 0 || len(prev.VisibleParticleEffectIDs) > 0 {
			if !internal.EqualEffectIDs(prev.VisibleParticleEffectIDs, s.VisibleParticleEffectIDs) {
				r.r.PushPlayerSetVisibleEffects(e, s.VisibleParticleEffectIDs)
			}
		}
	default:
		s := internal.GetEntityState(e)
		r.entityStatesMu.Lock()
		prev, ok := r.entityStates[e.H().UUID()]
		r.entityStates[e.H().UUID()] = s
		r.entityStatesMu.Unlock()

		if !ok {
			r.r.PushSetEntityNameTag(e, s.NameTag)
			return
		}

		if prev.NameTag != s.NameTag {
			r.r.PushSetEntityNameTag(e, s.NameTag)
		}
	}
}

func (r *RecorderViewer) ViewEntityAnimation(e world.Entity, a world.EntityAnimation) {
	// TODO: implement this
}

func (r *RecorderViewer) ViewParticle(pos mgl64.Vec3, p world.Particle) {
	switch p := p.(type) {
	case particle.BlockBreak:
		r.r.PushBlockBreakParticle(cube.PosFromVec3(pos), p.Block)
	case particle.PunchBlock:
		r.r.PushBlockPunchingParticle(cube.PosFromVec3(pos), p.Block, p.Face)
	default:
		r.r.PushGeneralParticle(pos, p)
	}
}

func (r *RecorderViewer) ViewSound(pos mgl64.Vec3, s world.Sound) {
	vec := cube.PosFromVec3(pos)
	switch s := s.(type) {
	case sound.BlockBreaking:
		r.r.PushBlockBreakingSound(vec, s.Block)
	case sound.BlockPlace:
		r.r.PushBlockPlaceSound(vec, s.Block)
	case sound.BucketEmpty:
		r.r.PushLiquidSound(vec, s.Liquid, false)
	case sound.BucketFill:
		r.r.PushLiquidSound(vec, s.Liquid, true)
	default:
		r.r.PushGeneralSound(pos, s)
	}
}

func (r *RecorderViewer) ViewBlockUpdate(pos cube.Pos, b world.Block, layer int) {
	switch layer {
	case 0:
		r.r.PushSetBlock(pos, b)
	case 1:
		if liq, ok := b.(world.Liquid); ok {
			r.r.PushSetLiquid(pos, liq)
		}
	}
}

func (r *RecorderViewer) ViewBlockAction(pos cube.Pos, a world.BlockAction) {
	switch a := a.(type) {
	case block.OpenAction:
		r.r.PushChestUpdate(pos, true)
	case block.CloseAction:
		r.r.PushChestUpdate(pos, false)
	case block.StartCrackAction:
		r.r.PushStartCrackBlock(pos, a.BreakTime)
	case block.StopCrackAction:
		r.r.PushStopCrackBlock(pos)
	case block.ContinueCrackAction:
		r.r.PushContinueCrackBlock(pos, a.BreakTime)
	}
}

func (r *RecorderViewer) ViewEmote(e world.Entity, emote uuid.UUID) {
	switch e := e.(type) {
	case *player.Player:
		r.r.PushEmote(e, emote)
	}
}

func (r *RecorderViewer) ViewSkin(e world.Entity) {
	// TODO: implement this
}
