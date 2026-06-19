package utils

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// networkOffset can be found at github.com\df-mc\dragonfly\server\player.(ptype).NetworkOffset()
const (
	PlayerHeight        = float64(1.8)
	PlayerWidth         = float64(0.6)
	DefaultSlipperiness = float64(0.6)
	SprintMovementMult  = float64(1.3)
	SprintJumpBoost     = float64(0.2)
	JumpSpeed           = float64(0.42)
	NetworkOffset       = float64(1.62)
	MomentumThreshold   = float64(0.003)
	GroundProbeOffset   = float64(0.003)
	HoriProbeOffset     = float64(0.003)
	Negligible          = float64(0.003)
)

// return > 0 if currentOffset is closer to 0 ,0 if equals, else < 0
func IsCloserToZero(currentOffset float64, leastOffset float64) float64 {
	return math.Abs(leastOffset) - math.Abs(currentOffset)
}

func DeltaIsZero(d mgl64.Vec3) bool {
	return d[0] == 0 && d[1] == 0 && d[2] == 0
}

func PlayerBBox(pos mgl64.Vec3) cube.BBox {
	halfW := PlayerWidth / 2
	return cube.Box(
		pos[0]-halfW,
		pos[1],
		pos[2]-halfW,
		pos[0]+halfW,
		pos[1]+PlayerHeight,
		pos[2]+halfW,
	)
}

func RotationToPitchAndYaw(r mgl32.Vec3) (yaw, pitch float32) {
	r64 := Mgl32Vec3Tomgl64Vec3(r)
	xz := math.Sqrt(math.Pow(r64[0], 2) + math.Pow(r64[2], 2))
	mag := math.Sqrt(math.Pow(xz, 2) + math.Pow(r64[1], 2))

	pitch64, yaw64 := 180/math.Pi, 180/math.Pi
	if xz > Negligible {
		yaw64 = math.Acos(r64[2]/xz) * 180 / math.Pi
	}
	if mag > Negligible {
		pitch64 = math.Acos(xz/mag) * 180 / math.Pi
	}
	pitch, yaw = float32(pitch64), float32(yaw64)
	return
}

func Mgl32Vec3Tomgl64Vec3(v mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{float64(v[0]), float64(v[1]), float64(v[2])}
}

func Mgl64Vec3Tomgl32Vec3(v mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v[0]), float32(v[1]), float32(v[2])}
}