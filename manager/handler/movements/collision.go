package movements

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

const stepHeight = 0.6

// getCollision resolves movement against blocks for the current velocity.
func (m *Movement) getCollision() collisionResult {
	//fmt.Printf("In block: %v\n", m.bboxIntersectsSolid(playerBBox(m.position)))
	pBBox := utils.PlayerBBox(m.playerPosBeforeVelocityApply())
	result := m.probeMovement(pBBox, m.velocity, &m.scratch.footOffsets)
	//result = m.tryStepUp(pBBox, result)
	return result
}

func (m *Movement) applyCollision(result collisionResult) {
	defer func ()  {
		fmt.Printf("Collision on tick %d: %+v\n", m.state.GStick(), result)
	}()
	if len(result.hittedAxis) == 0 {
		return
	}

	axis := result.oneExistAxis()
	offset := result.offsetOn(axis)
	// If offset equals velocity, travel is limited by speed rather than collision.
	if m.velocity[axis] == offset {
		//fmt.Println("m.velocity[axis] == offset")
		return
	}

	start := m.playerPosBeforeVelocityApply()
	otherAxes, reachable := utils.LineCoordAt(start, m.velocity, axis, start[axis]+offset)
	if !reachable {
		//fmt.Println("Not reachable")
		m.position[axis] = start[axis] + offset
		return
	}

	m.position = otherAxes
	for axis := range result.hittedAxis {
		//fmt.Printf("Velocity index on 0: %d\n", result.indices[i])
		m.velocity[axis] = 0
	}
}

func (m *Movement) probeMovement(pBBox cube.BBox, deltas mgl64.Vec3, out *axisOffsets) collisionResult {
	out.reset(deltas)
	if utils.DeltaIsZero(deltas) {
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
			offset, ok := planeOnCollide(pBBox, blockBox, [3]bool{xSolid, ySolid, zSolid}, deltas)
			if !ok{
				continue
			}
			out[offset.axis].consider(offset.offset, blockBox)
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
	result := collisionResult{offsets: *offsets, hittedAxis: make(map[int]struct{}, 3)}
	for axis, ratio := range utils.MinOffset(offsets.offsetArr(), deltas) {
		if ratio == 1 {
			return result
		}
		result.hittedAxis[axis] = struct{}{}
	}
	return result
}

func (m *Movement) tryStepUp(pBBox cube.BBox, result collisionResult) collisionResult {
	if _, ok := result.hittedAxis[1]; !(m.onGround && m.velocity[1] == 0 && !ok) {
		return result
	}
	if len(result.hittedAxis) == 0 {
		return result
	}

	var height float64
	allTravelLimited := true
	for axis := range result.hittedAxis {
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
	for axis := range result.hittedAxis{
		if m.velocity[axis] == result.offsetOn(axis) {
			continue
		}
		if utils.IsCloserToZero(stepResult.offsetOn(axis), result.offsetOn(axis)) <= 0 {
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
