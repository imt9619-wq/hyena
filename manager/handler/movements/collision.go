package movements

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

const stepHeight = 0.6

// getCollision resolves movement against blocks for the current velocity.
func (m *Movement) getCollision() collisionResult {
	//fmt.Printf("In block: %v\n", m.bboxIntersectsSolid(playerBBox(m.position)))
	pBBox := playerBBox(m.playerPosBeforeVelocityApply())
	result := m.probeMovement(pBBox, m.velocity, &m.scratch.footOffsets)
	return result //m.tryStepUp(pBBox, result)
}

func (m *Movement) applyCollision(result collisionResult) {
	defer func ()  {
		fmt.Printf("Collision on tick %d: %+v\n", m.state.GStick(), result)
	}()
	if result.nIndices == 0 {
		return
	}

	axis := result.indices[0]
	offset := result.offsetOn(axis)
	// If offset equals velocity, travel is limited by speed rather than collision.
	if m.velocity[axis] == offset {
		//fmt.Println("m.velocity[axis] == offset")
		return
	}

	start := m.playerPosBeforeVelocityApply()
	otherAxes, reachable := lineCoordAt(start, m.velocity, axis, start[axis]+offset)
	if !reachable {
		//fmt.Println("Not reachable")
		m.position[axis] = start[axis] + offset
		return
	}

	m.position = otherAxes
	for i := 0; i < result.nIndices; i++ {
		//fmt.Printf("Velocity index on 0: %d\n", result.indices[i])
		m.velocity[result.indices[i]] = 0
	}
}

func (m *Movement) probeMovement(pBBox cube.BBox, deltas mgl64.Vec3, out *axisOffsets) collisionResult {
	out.reset(deltas)
	if deltaIsZero(deltas) {
		return collisionResult{offsets: *out}
	}

	bm := m.state.BlockMap()
	xFace, yFace, zFace := cube.FaceWest, cube.FaceDown, cube.FaceNorth
	if deltas[0] > 0 {
		xFace = cube.FaceEast
	}
	if deltas[1] > 0 {
		yFace = cube.FaceUp
	}
	if deltas[2] > 0 {
		zFace = cube.FaceSouth
	}

	for pos := range m.sweptBlockPositions(pBBox, deltas) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok {
			continue
		}
		blockBoxes := bboxes(model, pos, bm)
		xSolid := model.FaceSolid(pos, xFace, bm)
		ySolid := model.FaceSolid(pos, yFace, bm)
		zSolid := model.FaceSolid(pos, zFace, bm)
		for _, blockBox := range blockBoxes {
			//fmt.Printf("Checked BBoxes: %+v\n", blockBox)
			offset, ok := planeOnCollide(pBBox, blockBox, [3]bool{xSolid, ySolid, zSolid}, deltas)
			if !ok{
				continue
			}
			out[offset.plane].consider(offset.offset, blockBox)
		}
	}
	return m.earliestAxes(out, deltas)
}

func bboxes(model world.BlockModel, pos cube.Pos, s world.BlockSource) []cube.BBox{
	blockBoxes := model.BBox(pos, s)
	for i, bbox := range blockBoxes{
		blockBoxes[i] = bbox.Translate(pos.Vec3())
	}
	return blockBoxes
}

func (m *Movement) earliestAxes(offsets *axisOffsets, deltas mgl64.Vec3) collisionResult {
	var ratios [3]float64
	for i := range ratios {
		ratios[i] = 1
		if deltas[i] != 0 {
			ratios[i] = offsets[i].offset / deltas[i]
		}
	}
	minRatio := min(ratios[0], ratios[1], ratios[2])
	result := collisionResult{offsets: *offsets}
	if minRatio == 1{
		return result
	}
	for i := range ratios {
		if ratios[i] == minRatio {
			result.indices[result.nIndices] = i
			result.nIndices++
		}
	}
	return result
}

func (m *Movement) tryStepUp(pBBox cube.BBox, result collisionResult) collisionResult {
	if !(m.onGround && m.velocity[1] == 0 && !result.hitsAxis(1)) {
		return result
	}
	if result.nIndices == 0 {
		return result
	}

	var height float64
	allTravelLimited := true
	for i := 0; i < result.nIndices; i++ {
		axis := result.indices[i]
		offset := result.offsetOn(axis)
		if m.velocity[axis] == offset {
			continue
		}

		allTravelLimited = false
		for _, blockBox := range result.blocksOn(axis) {
			currentHeight := blockBox.Max()[1] - pBBox.Min()[1]
			if currentHeight > height {
				height = currentHeight
			}
			if currentHeight > stepHeight {
				return result
			}
		}
	}
	if height == 0 || allTravelLimited {
		return result
	}

	raisedBBox := pBBox.Translate(mgl64.Vec3{0, height, 0})
	if m.bboxIntersectsSolid(raisedBBox) {
		return result
	}

	stepResult := m.probeMovement(raisedBBox, m.velocity, &m.scratch.stepOffsets)
	for i := 0; i < result.nIndices; i++ {
		axis := result.indices[i]
		if m.velocity[axis] == result.offsetOn(axis) {
			continue
		}
		if isCloserToZero(stepResult.offsetOn(axis), result.offsetOn(axis)) <= 0 {
			return result
		}
	}

	m.position[1] += height
	return m.earliestAxes(&stepResult.offsets, m.velocity)
}

func (m *Movement) bboxIntersectsSolid(pBBox cube.BBox) bool {
	bm := m.state.BlockMap()
	for _, pos := range m.blockPositionsInBBox(pBBox) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok {
			continue
		}
		for _, blockBox := range bboxes(model, pos, bm) {
			if pBBox.IntersectsWith(blockBox) {
				return true
			}
		}
	}
	return false
}
