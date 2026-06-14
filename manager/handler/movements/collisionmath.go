package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// This function will get all the block pos that the player BBox will came accoss from the position of last
// tick to the future position after applying the velocity on the position,
func (m *Movement) intersection(pBBox cube.BBox, deltas mgl64.Vec3) map[cube.Pos]struct{} {
	clear(m.cc.intersectBlocks)
	pCorners := pBBox.Corners()
	velocity := deltas

	for _, corner := range pCorners {
		for index, val := range corner {
			if velocity[index] == 0 {
				continue
			}
			for _, point := range floorFloatBetweenAB(val, val+velocity[index], &m.cc.floorPointsScratch) {
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
				m.cc.intersectBlocks[Mgl64Vec3ToCubePos(newVec3)] = struct{}{}
			}
		}
	}
	return m.cc.intersectBlocks
}

func (m *Movement) blockPositionsInBBox(bbox cube.BBox) []cube.Pos {
	min := bbox.Min()
	max := bbox.Max()
	m.cc.blockPosScratch = m.cc.blockPosScratch[:0]
	for x := int(math.Floor(min[0])); x <= int(math.Floor(max[0])); x++ {
		for y := int(math.Floor(min[1])); y <= int(math.Floor(max[1])); y++ {
			for z := int(math.Floor(min[2])); z <= int(math.Floor(max[2])); z++ {
				m.cc.blockPosScratch = append(m.cc.blockPosScratch, cube.Pos{x, y, z})
			}
		}
	}
	return m.cc.blockPosScratch
}

// since a block position is gonna be a 3 integer array, we just need to get the integer in range of the integer
// of the player last tick position to the future tick position, for example, if a player when from [2.5, 0, 1.1]
// to [-5.1, 2, 0.6](pretty big changes for a tick, it is rare just for demostration) then the integer x,y,z in
// range of the last tick to the future tick is -5,-4,-3,-2,-1,0,1,2 for x. 0,1,2 for y. 1 for z.
// Then we can input all the value of the axis to threeDLine() to get all the points the player BBox will come
// accross on its way to the future tick position, since it is possible to have mutiple point of the same point
// of one axis of Movement, we need the changes of integer from all axis instead of just inputing plotting one point
// of an axis to the threeDLine() as we all one get one return value, we designed the threeDLine is this way as we
// saw it as a simplier aroppoch
func floorFloatBetweenAB(a float64, b float64, scratch *[]float64) []float64 {
	if a > b {
		a, b = b, a
	}
	*scratch = (*scratch)[:0]
	for i := math.Floor(a); i <= b; i++ {
		*scratch = append(*scratch, i)
	}
	return *scratch
}

// this function will take an inputPointIndex, 0 for x, 1 for y, 2 for z, then the
// value of that index, we also need a point that is known to be on the line(i) and of
// course the vector or slope of the line(direction), the returning points index are assending
// from left to right, so it can output the value of index in order of 0,2 0,1 or 1,2 , bool
// will return false if that point is impossible to reach, by inputting a value of a index we can
// calualate the whole pair of coordinate meaning we can get the y,z by inputting x, where the pair is one
// of the point that the player will come accoss on its path from the last tick position to the future tick(when
// we are saying future tick, that doesnt nessary means that the player will be in that position in the next tick,
// it is only the case if the player havent collision with any block on its way)
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