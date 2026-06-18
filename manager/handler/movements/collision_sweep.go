package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// sweptBlockPositions returns block positions the player bbox crosses while moving by deltas.
func (m *Movement) sweptBlockPositions(pBBox cube.BBox, deltas mgl64.Vec3) map[cube.Pos]struct{} {
	clear(m.scratch.sweepBlocks)
	corners := pBBox.Corners()

	for _, corner := range corners {
		for axis, start := range corner {
			if deltas[axis] == 0 {
				continue
			}
			for _, plane := range floatPlanesBetween(start, start+deltas[axis], &m.scratch.floorPointsScratch) {
				axisPair, ok := lineCoordAt(corner, deltas, axis, plane)
				if !ok {
					break
				}
				m.scratch.sweepBlocks[Mgl64Vec3ToCubePos(axisPair)] = struct{}{}
			}
		}
	}
	return m.scratch.sweepBlocks
}

func (m *Movement) blockPositionsInBBox(bbox cube.BBox) []cube.Pos {
	min := bbox.Min()
	max := bbox.Max()
	m.scratch.blockPosScratch = m.scratch.blockPosScratch[:0]
	for x := int(math.Floor(min[0])); x <= int(math.Floor(max[0])); x++ {
		for y := int(math.Floor(min[1])); y <= int(math.Floor(max[1])); y++ {
			for z := int(math.Floor(min[2])); z <= int(math.Floor(max[2])); z++ {
				m.scratch.blockPosScratch = append(m.scratch.blockPosScratch, cube.Pos{x, y, z})
			}
		}
	}
	return m.scratch.blockPosScratch
}

// floatPlanesBetween returns each integer boundary crossed between a and b.
func floatPlanesBetween(a, b float64, scratch *[]float64) []float64 {
	if a > b {
		a, b = b, a
	}
	*scratch = (*scratch)[:0]
	for i := math.Floor(a); i <= b; i++ {
		*scratch = append(*scratch, i)
	}
	return *scratch
}

// lineCoordAt returns the coordinates on the whole vector point where a movement line
// from origin with direction crosses the given plane on fixedAxis.
func lineCoordAt(origin, direction mgl64.Vec3, fixedAxis int, plane float64) (mgl64.Vec3, bool) {
	if direction[fixedAxis] == 0 {
		return mgl64.Vec3{}, false
	}
	t := (plane - origin[fixedAxis]) / direction[fixedAxis]
	var out mgl64.Vec3
	for axis, val := range origin {
		if axis == fixedAxis {
			out[axis] = plane
			continue
		}
		out[axis] = val + t*direction[axis]
	}
	return out, true
}
