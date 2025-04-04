package action

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
)

func EncodeYaw16(yaw float32) uint16 {
	clamped := mgl32.Clamp(yaw, -180, 180)
	return uint16(((clamped + 180) / 360.0) * 65535.0)
}

func DecodeYaw16(encoded uint16) float32 {
	return (float32(encoded) / 65535.0 * 360.0) - 180.0
}

func EncodePitch16(pitch float32) uint16 {
	clamped := mgl32.Clamp(pitch, -90, 90)
	return uint16(((clamped + 90) / 180.0) * 65535.0)
}

func DecodePitch16(encoded uint16) float32 {
	return (float32(encoded) / 65535.0 * 180.0) - 90.0
}

func EncodeRotation16(rot cube.Rotation) (uint16, uint16) {
	return EncodeYaw16(float32(rot[0])), EncodeYaw16(float32(rot[1]))
}

func DecodeRotation16(yaw, pitch uint16) cube.Rotation {
	return cube.Rotation{
		float64(DecodeYaw16(yaw)),
		float64(DecodePitch16(pitch)),
	}
}
