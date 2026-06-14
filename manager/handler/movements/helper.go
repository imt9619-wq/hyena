package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// return > 0 if currentOffset is closer to 0 ,0 if equals, else < 0
func isCloserToZero(currentOffset float64, leastOffset float64) float64 {
	return math.Abs(leastOffset) - math.Abs(currentOffset)
}

func deltaIsZero(d mgl64.Vec3) bool {
	return d[0] == 0 && d[1] == 0 && d[2] == 0
}

func playerBBox(pos mgl64.Vec3) cube.BBox {
	halfW := playerWidth / 2
	return cube.Box(
		pos[0]-halfW,
		pos[1],
		pos[2]-halfW,
		pos[0]+halfW,
		pos[1]+playerHeight,
		pos[2]+halfW,
	)
}

func RotationToPitchAndYaw(r mgl32.Vec3) (yaw, pitch float32) {
	r64 := mgl32Vec3Tomgl64Vec3(r)
	xz := math.Sqrt(math.Pow(r64[0], 2) + math.Pow(r64[2], 2))
	mag := math.Sqrt(math.Pow(xz, 2) + math.Pow(r64[1], 2))

	pitch64, yaw64 := 180/math.Pi, 180/math.Pi
	if xz > negligible {
		yaw64 = math.Acos(r64[2]/xz) * 180 / math.Pi
	}
	if mag > negligible {
		pitch64 = math.Acos(xz/mag) * 180 / math.Pi
	}
	pitch, yaw = float32(pitch64), float32(yaw64)
	return
}

func mgl32Vec3Tomgl64Vec3(v mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{float64(v[0]), float64(v[1]), float64(v[2])}
}

func mgl64Vec3Tomgl32Vec3(v mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v[0]), float32(v[1]), float32(v[2])}
}

func Mgl64Vec3ToCubePos(v mgl64.Vec3) cube.Pos {
	return cube.Pos{
		int(math.Floor(v[0])),
		int(math.Floor(v[1])),
		int(math.Floor(v[2])),
	}
}

