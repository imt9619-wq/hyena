package sim

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

func (m *Movement) intersection(pBBox cube.BBox, deltas mgl32.Vec3) map[cube.Pos]struct{} {
	pCorners := pBBox.Corners()
	velocity := mgl32Vec3ToMgl64Vec3(deltas)

	intersectedBlocks := make(map[cube.Pos]struct{}, 10)
	for _, corner := range pCorners {
		for index, val := range corner {
			for _, point := range floorFloatBetweenAB(val, val+velocity[index]) {
				other2Point, exist := threeDLine(corner, velocity, index, point)
				if !exist {
					break
				}
				var newVec3 mgl64.Vec3
				currentOther2PointIndex := 0
				for i := 0; i <= 2; i++ {
					if i != index {
						newVec3[i] = other2Point[currentOther2PointIndex]
						currentOther2PointIndex++
						continue
					}
					newVec3[i] = point
				}
				intersectedBlocks[mgl64Vec3ToCubePos(newVec3)] = struct{}{}
			}
		}
	}
	return intersectedBlocks
}

func floorFloatBetweenAB(a float64, b float64) []float64 {
	if a > b {
		a, b = b, a
	}
	ceilDistance := int(math.Ceil(b - a))
	pointsInAB := make([]float64, 0, ceilDistance)
	for i := math.Floor(a); i <= b; i++ {
		pointsInAB = append(pointsInAB, i)
	}
	return pointsInAB
}

func mgl64Vec3ToCubePos(v mgl64.Vec3) cube.Pos {
	return cube.Pos{
		int(math.Floor(v[0])),
		int(math.Floor(v[1])),
		int(math.Floor(v[2])),
	}
}

func threeDLine(i mgl64.Vec3, direction mgl64.Vec3, inputPointIndex int, inputPointValue float64) (mgl64.Vec2, bool) {
	var outputPointsPair mgl64.Vec2
	if direction[inputPointIndex] == 0 {
		return outputPointsPair, false
	}
	t := (inputPointValue - i[inputPointIndex]) / direction[inputPointIndex]

	nextIndex := 0
	for index, val := range i {
		if index == inputPointIndex {
			continue
		}
		outputPointsPair[nextIndex] = val + t*direction[index]
		nextIndex++
	}
	return outputPointsPair, true
}

func mgl32Vec3ToMgl64Vec3(v mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{float64(v[0]), float64(v[1]), float64(v[2])}
}
