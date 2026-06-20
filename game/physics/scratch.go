package physics

import (
	"slices"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

type aabbToGrid struct{
	rmu  *sync.RWMutex
    grid map[cube.BBox][]mgl64.Vec3
}

var toGrid = &aabbToGrid{
	rmu: &sync.RWMutex{},
	grid: make(map[cube.BBox][]mgl64.Vec3, 10),
}

func getGrid(aabb cube.BBox, gridScratch *[]mgl64.Vec3) bool{
	aabbOffset := aabb.Min()
	toGrid.rmu.RLock()
	grids, ok := toGrid.grid[aabb.Translate(aabbOffset.Mul(-1))]
	if !ok{
		toGrid.rmu.RUnlock()
		return false
	}
	*gridScratch = (*gridScratch)[:0]
	for _, g := range grids {
		*gridScratch = append(*gridScratch, g.Add(aabbOffset))
	}
	toGrid.rmu.RUnlock()
	return true
}

func insertGrid(aabb cube.BBox, gridScratch *[]mgl64.Vec3){
	aabbOffset := aabb.Min()
	toGrid.rmu.RLock()
	_, ok := toGrid.grid[aabb.Translate(aabbOffset.Mul(-1))]
	toGrid.rmu.RUnlock()
	if ok{
		return
	}
	transLatedGrids := slices.Clone(*gridScratch)
	for axis, plane := range transLatedGrids{
		transLatedGrids[axis] = plane.Sub(aabbOffset)
	}
	toGrid.rmu.Lock()
	toGrid.grid[aabb.Translate(aabbOffset.Mul(-1))] = transLatedGrids
	toGrid.rmu.Unlock()
}

type phyScratch struct {
	blockInPath map[cube.Pos]struct{}
	aabbGrid     []mgl64.Vec3
	offsets      *axisOffsets
}

func newScratch() *phyScratch{
	p := &phyScratch{
		blockInPath: make(map[cube.Pos]struct{}, 128),
		aabbGrid: make([]mgl64.Vec3, 0, 16),
		offsets: &axisOffsets{},
	}
	return p
}

// sweptBlockPositions returns block positions the player bbox crosses while moving by deltas.
func (p *phyScratch) sweptBlockPositions(aabb cube.BBox, deltas mgl64.Vec3) map[cube.Pos]struct{} {
	clear(p.blockInPath)
	corners := p.aabbGrids(aabb.Grow(0.2))
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
				p.blockInPath[cube.PosFromVec3(axisPair)] = struct{}{}
			}
		}
	}
	return p.blockInPath
}

func (p *phyScratch) aabbGrids(aabb cube.BBox) []mgl64.Vec3{
	if getGrid(aabb, &p.aabbGrid){
		return p.aabbGrid
	}
	p.aabbGrid = p.aabbGrid[:0]
	for x := range floatIntBetween(aabb.Min()[0], aabb.Max()[0]){
		for y := range floatIntBetween(aabb.Min()[1], aabb.Max()[1]){
			for z := range floatIntBetween(aabb.Min()[2], aabb.Max()[2]){
				p.aabbGrid = append(p.aabbGrid, mgl64.Vec3{x, y, z})
			}
		}
	}
	insertGrid(aabb, &p.aabbGrid)
	return p.aabbGrid
}

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
	return [3]float64{a[0].offset, a[1].offset, a[2].offset}
}

func (a *axisOffsets) reset(deltas mgl64.Vec3) {
	for i := range a {
		a[i].offset = deltas[i]
		a[i].blocks = a[i].blocks[:0]
	}
}

func (a *axisOffsets) considerOffsets(self, nearby cube.BBox, solid [3]bool, deltas mgl64.Vec3){
	for axis, isSolid := range solid{
		if isSolid{
			if offset, ok := AOffset(self, nearby, axis, deltas); ok{
				a[axis].consider(offset, nearby)
			}
		}
	}
}

func (s *StateInWorld) ScratchOffset() *axisOffsets{
	return s.scratch.offsets
}