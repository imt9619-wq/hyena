package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

// axisOffset is the closest allowed travel distance on one axis before hitting a block.
type axisOffset struct {
	offset float64
	blocks []cube.BBox
}

func (o *axisOffset) consider(candidate float64, block cube.BBox) {
	if utils.IsCloserToZero(candidate, o.offset) > 0 {
		o.blocks = o.blocks[:0]
		o.offset = candidate
	}
	if candidate == o.offset {
		o.blocks = append(o.blocks, block)
	}
}

// axisOffsets holds per-axis collision results for a single movement probe.
type axisOffsets [3]axisOffset

func (a *axisOffsets) offsetArr() [3]float64{
	var arr [3]float64
	for i, offset := range a{
		arr[i] = offset.offset
	}
	return arr
}

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
	hittedAxis map[int]struct{}
}

func (r collisionResult) oneExistAxis() int{
	for axis := range r.hittedAxis{
		return axis
	}
	return -1
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
	footOffsets        axisOffsets
	stepOffsets        axisOffsets
}

func newCollisionScratch() *collisionScratch {
	return &collisionScratch{
		sweepBlocks: make(map[cube.Pos]struct{}, 16),
	}
}

type collidePlane struct{
	axis  int
    offset float64
}

func planeOnCollide(self, nearby cube.BBox, solid [3]bool, delta mgl64.Vec3) (collidePlane, bool){
	var offset float64
	var collidePlane collidePlane
	radio := mgl64.MaxValue
	exist := false
	for axis, plane := range delta{
		if !solid[axis] || plane == 0{
			continue
		}
		offset = plane
		if plane > 0 && self.Max()[axis] <= nearby.Min()[axis]{
			offset = min(nearby.Min()[axis] - self.Max()[axis], plane)
		}
		if plane < 0 && self.Min()[axis] >= nearby.Max()[axis]{
			offset = max(nearby.Max()[axis] - self.Min()[axis], plane)
		}
		
		if offset != plane && !utils.OutOfPlane(self.Translate(delta.Mul(offset/plane)), nearby, axis){
			if offset/plane < radio{
				collidePlane.axis, collidePlane.offset = axis, offset
				radio = offset/plane
				exist = true
				if mgl64.FloatEqualThreshold(collidePlane.offset, 0, utils.Negligible){
					collidePlane.offset = 0
					break
				}
			}
		}
	}
	return collidePlane, exist
}
