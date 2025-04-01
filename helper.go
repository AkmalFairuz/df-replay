package replay

import (
	"github.com/akmalfairuz/df-replay/action"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func vec64To32(v mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v[0]), float32(v[1]), float32(v[2])}
}

func cubeToBlockPos(pos cube.Pos) protocol.BlockPos {
	return protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}
}

func skinToAction(playerID uint32, sk skin.Skin) *action.PlayerSkin {
	return &action.PlayerSkin{
		PlayerID:        playerID,
		SkinWidth:       uint32(sk.Bounds().Dx()),
		SkinHeight:      uint32(sk.Bounds().Dy()),
		SkinData:        sk.Pix,
		HasCape:         len(sk.Cape.Pix) > 0,
		CapeWidth:       uint32(sk.Cape.Bounds().Dx()),
		CapeHeight:      uint32(sk.Cape.Bounds().Dy()),
		CapeData:        sk.Cape.Pix,
		GeometryName:    sk.ModelConfig.Default,
		HasGeometryData: len(sk.Model) > 0,
		GeometryData:    sk.Model,
	}
}
