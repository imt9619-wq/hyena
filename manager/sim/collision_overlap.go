package sim

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func blockPositionsInBBox(bbox cube.BBox) []cube.Pos {
	min := bbox.Min()
	max := bbox.Max()
	positions := make([]cube.Pos, 0, 8)
	for x := int(math.Floor(min[0])); x <= int(math.Floor(max[0])); x++ {
		for y := int(math.Floor(min[1])); y <= int(math.Floor(max[1])); y++ {
			for z := int(math.Floor(min[2])); z <= int(math.Floor(max[2])); z++ {
				positions = append(positions, cube.Pos{x, y, z})
			}
		}
	}
	return positions
}

func (m *Movement) bboxIntersectsSolid(pBBox cube.BBox) bool {
	world := m.session.BlockMap
	for _, pos := range blockPositionsInBBox(pBBox) {
		model, ok := world.BlockModel(pos, 0)
		if !ok {
			continue
		}
		for _, bBBox := range model.BBox(pos, world) {
			if pBBox.IntersectsWith(bBBox) {
				return true
			}
		}
	}
	return false
}
