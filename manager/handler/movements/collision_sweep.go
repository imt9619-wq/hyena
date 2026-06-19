package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

// sweptBlockPositions returns block positions the player bbox crosses while moving by deltas.
func (m *Movement) sweptBlockPositions(pBBox cube.BBox, deltas mgl64.Vec3) map[cube.Pos]struct{} {
	clear(m.scratch.sweepBlocks)
	corners := pBBox.Grow(0.2).Corners()

	for _, corner := range corners {
		for axis, start := range corner {
			if deltas[axis] == 0 {
				continue
			}
			for plane := range utils.FloorFloatBetween(start, start+deltas[axis]) {
				axisPair, ok := utils.LineCoordAt(corner, deltas, axis, plane)
				if !ok {
					break
				}
				m.scratch.sweepBlocks[cube.PosFromVec3(axisPair)] = struct{}{}
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

