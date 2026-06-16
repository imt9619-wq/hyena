package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// axisOffset is the closest allowed travel distance on one axis before hitting a block.
type axisOffset struct {
	offset float64
	blocks []cube.BBox
}

func (o *axisOffset) consider(candidate float64, block cube.BBox) {
	if isCloserToZero(candidate, o.offset) > 0 {
		o.blocks = o.blocks[:0]
		o.offset = candidate
	}
	if candidate == o.offset {
		o.blocks = append(o.blocks, block)
	}
}

// axisOffsets holds per-axis collision results for a single movement probe.
type axisOffsets [3]axisOffset

func (a *axisOffsets) reset(deltas mgl64.Vec3) {
	for i := range a {
		a[i].offset = deltas[i]
		a[i].blocks = a[i].blocks[:0]
	}
}

// collisionResult is the outcome of probing movement against blocks: per-axis offsets
// and which axis(es) would be hit first when moving by deltas.
type collisionResult struct {
	offsets   axisOffsets
	indices   [3]int
	nIndices  int
}

func (r collisionResult) hitsAxis(axis int) bool {
	for i := 0; i < r.nIndices; i++ {
		if r.indices[i] == axis {
			return true
		}
	}
	return false
}

func (r collisionResult) offsetOn(axis int) float64 {
	return r.offsets[axis].offset
}

func (r collisionResult) blocksOn(axis int) []cube.BBox {
	return r.offsets[axis].blocks
}

// collisionScratch holds reusable buffers for block queries within a tick.
type collisionScratch struct {
	sweepBlocks        map[cube.Pos]struct{}
	blockPosScratch    []cube.Pos
	floorPointsScratch []float64
	footOffsets        axisOffsets
	stepOffsets        axisOffsets
}

func newCollisionScratch() *collisionScratch {
	return &collisionScratch{
		sweepBlocks: make(map[cube.Pos]struct{}, 16),
	}
}
