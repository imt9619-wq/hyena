package utils

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// return > 0 if currentOffset is closer to 0 ,0 if equals, else < 0
func IsCloserToZero(currentOffset float64, leastOffset float64) float64 {
	return math.Abs(leastOffset) - math.Abs(currentOffset)
}

func DeltaIsZero(d mgl64.Vec3) bool {
	return d[0] == 0 && d[1] == 0 && d[2] == 0
}

type BBoxFunc func(mgl64.Vec3) cube.BBox

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

func Mgl32Vec3Tomgl64Vec3(v mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{float64(v[0]), float64(v[1]), float64(v[2])}
}

func Mgl64Vec3Tomgl32Vec3(v mgl64.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v[0]), float32(v[1]), float32(v[2])}
}

func RoundFloat(val float64, precision uint) float64 {
    ratio := math.Pow(10, float64(precision))
    return math.Round(val * ratio) / ratio
}

func RoundVecTo5Decimal(delta mgl64.Vec3) mgl64.Vec3{
	for axis, plane := range delta{
		delta[axis] = math.Round(plane*100000) / 100000
	}
	return delta
}

func RemoveDeltaEpsilon(delta mgl64.Vec3) mgl64.Vec3{
	for axis, plane := range delta{
		if mgl64.FloatEqualThreshold(plane, 0, Epsilon){
			delta[axis] = 0
		}
	}
	return delta
}

func xzSpeed(v mgl32.Vec3) float32 {
	return float32(math.Sqrt(math.Pow(float64(v[0]), 2) + math.Pow(float64(v[2]), 2)))
}

func sinNCosOfSpeed(velocity mgl32.Vec3) (sinD, cosD float32) {
	speed := xzSpeed(velocity)
	xVel := velocity[0]
	zVel := velocity[2]

	sinD = float32(0)
	cosD = float32(1)
	if speed > 0.003 {
		sinD = xVel / speed
		cosD = zVel / speed
	}
	return
}

func SpeedToVelocity(velocity mgl32.Vec3, speed float32) mgl32.Vec3{
	sinD, cosD := sinNCosOfSpeed(velocity)
	velocity[0] = speed*sinD
	velocity[2] = speed*cosD
	return velocity
}
