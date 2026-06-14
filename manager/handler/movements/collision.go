package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

type leastBoxOffset struct {
	leastOffset       float64
	leastOffsetBBoxes []cube.BBox
}

type offsetAxis [3]leastBoxOffset

type collideOffsets struct {
	closestBox offsetAxis
	indies     [3]int
	nIndies    int
}

type collisionCache struct {
	intersectBlocks    map[cube.Pos]struct{}
	blockPosScratch    []cube.Pos
	floorPointsScratch []float64
	footScratch        offsetAxis
	stepScratch        offsetAxis
}

func newCollisionCache() *collisionCache {
	cc := &collisionCache{
		intersectBlocks: make(map[cube.Pos]struct{}, 16),
	}
	return cc
}

func (m *Movement) applyCollision(c collideOffsets) {
	if c.nIndies == 0 {
		return
	}

	index := c.indies[0]
	offset := c.closestBox[index].leastOffset
	// if m.velocity[index] == offset, that means the player arent colliding, it's just that all 
	// they can travel with that speed, in that case the player is limited by velocity instead of block bbox
	if m.velocity[index] == offset {
		return
	}

	pairs, reachable := threeDLine(m.playerPosBeforeVelocityApply(), 
								   m.velocity, 
								   index, 
								   m.posWithOffset(index, offset))
	// happens when the velocity is 0, and if so, then position is the same as beforeTick anyway
	if !reachable{
		m.position[index] = m.posWithOffset(index, offset)
		return
	}

	currentPair := 0
	for i := 0; i < 3; i++ {
		if i == index {
			m.position[i] = m.posWithOffset(index, offset)
			continue
		}
		m.position[i] = pairs[currentPair]
	}
	
	for i := 0; i < c.nIndies; i++ {
		m.velocity[c.indies[i]] = 0
	}
}	

func (m *Movement) posWithOffset(index int, offset float64) float64 {
	return m.playerPosBeforeVelocityApply()[index] + offset
}

// cheakCollision will check if the player will collision with a block after apply velocity on its
// position, if collision player velocity towards the axis where the player BBox and block BBox
// collided will be set to 0, the position will also be set the position where the player BBox is
// just touching the block BBox in the axis that they collided
func (m *Movement) getCollision() collideOffsets {
	pBBoxBeforeTick := playerBBox(m.playerPosBeforeVelocityApply())
	// first walk path as usual
	c := m.walkPath(pBBoxBeforeTick)
	// then try to walk up stairs
	c = m.walkUpStairs(pBBoxBeforeTick, c)
	return c
}

func (m *Movement) walkPath(pBBox cube.BBox) collideOffsets {
	return m.offsetsBeforeCollide(pBBox, m.velocity, &m.cc.footScratch)
}

func (a *offsetAxis) reset(deltas mgl64.Vec3) {
	for i := 0; i < 3; i++ {
		a[i].leastOffset = deltas[i]
		a[i].leastOffsetBBoxes = a[i].leastOffsetBBoxes[:0]
	}
}

func (c collideOffsets) collideIndiesContains(axis int) bool {
	for i := 0; i < c.nIndies; i++ {
		if c.indies[i] == axis {
			return true
		}
	}
	return false
}

func (m *Movement) walkUpStairs(pBBox cube.BBox, c collideOffsets) collideOffsets {
	// while players is walking stair, they shouldn't be jumping or falling or on air
	// and if the player is going to collide with a ceiling
	if !(m.onGround && m.velocity[1] == 0 && !c.collideIndiesContains(1)) {
		return c
	}
	// shouldn't happen at the first place
	if c.nIndies == 0 {
		return c
	}
	
	var stepHeight float64
	allOffsetEqualsVelocity := true
	for i := 0; i < c.nIndies; i++ {
		index := c.indies[i]
		offset := c.closestBox[index].leastOffset
		// return if not colliding with any block as a stair is still a bbox
		if m.velocity[index] == offset {
			continue
		}

		allOffsetEqualsVelocity = false
		bboxWithOffset := c.closestBox[index].leastOffsetBBoxes	
		for _, bBBox := range bboxWithOffset {
			currentStepHeight := bBBox.Max()[1] - pBBox.Min()[1]
			if currentStepHeight > stepHeight {
				stepHeight = currentStepHeight
			}
			// too high to walk up
			if currentStepHeight > 0.6 {
				return c
			}
		}
	}
	// shouldn't happen at the first place
	if stepHeight == 0 || allOffsetEqualsVelocity {
		return c
	}

	// return if ceiling is not high enough for walking up stairs
	raisedBBox := pBBox.Translate(mgl64.Vec3{0, stepHeight, 0})
	if m.bboxIntersectsSolid(raisedBBox) {
		return c
	}

	// m.velocity[1] is 0 as we returned the non zero m.velocity[1] case already
	stepC := m.offsetsBeforeCollide(raisedBBox, m.velocity, &m.cc.stepScratch)
	for i := 0; i < c.nIndies; i++ {
		index := c.indies[i]
		offset := c.closestBox[index].leastOffset
		if m.velocity[index] == offset {
			continue
		}
		// when stepping up a stair we expect the offset to be further to 0 than c.indies[0] is to 0
		// or else we are just trying to walk onto a wall while floating
		if isCloserToZero(stepC.closestBox[index].leastOffset, offset) <= 0 {
			return c
		}
	}
	// velocity[1] is 0 so we dont have to use playerPosBeforeTick
	m.position[1] += stepHeight
	return m.firstCollideAxis(&stepC.closestBox, m.velocity)
}

// will return the expected final delta on each axis of the player last tick position to its future tick,
// most of the time, only one offset on an axis is gonna be applied as the player will reach that axis of
// the bbox before the another two and if so the player will be stopped on that point already so the player
// wont actually be able to reach the another two axis, although there is some extreme case where offset to
// velocity radio is the same for two axis(or even three) where then the player will reach the two axis at
// the same time, therefore we are returning a slice instead of just a float for those cases(int is for
// index, float is the actual offset)
func (m *Movement) firstCollideAxis(closestBox *offsetAxis, deltas mgl64.Vec3) collideOffsets {
	var radios [3]float64
	for i := 0; i < 3; i++ {
		radios[i] = 1
		if deltas[i] != 0 {
			radios[i] = closestBox[i].leastOffset / deltas[i]
		}
	}
	minRadio := min(radios[0], radios[1], radios[2])
	c := collideOffsets{closestBox: *closestBox}
	for i := 0; i < 3; i++ {
		if radios[i] == minRadio {
			c.indies[c.nIndies] = i
			c.nIndies++
		}
	}
	return c
}


// return the least offsets(distance between bBBox and pBBox in some sense) of each axis
func (m *Movement) offsetsBeforeCollide(pBBoxBeforeTick cube.BBox, deltas mgl64.Vec3, out *offsetAxis) collideOffsets {
	out.reset(deltas)
	if deltaIsZero(deltas) {
		return collideOffsets{closestBox: *out}
	}

	bm := m.state.BlockMap()
	xSolid, ySolid, zSolid := true, true, true
	xOffset, yOffset, zOffset := out[0].leastOffset, out[1].leastOffset, out[2].leastOffset
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

	for pos := range m.intersection(pBBoxBeforeTick, deltas) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok {
			continue
		}
		bBBoxs := model.BBox(pos, bm)
		xSolid = model.FaceSolid(pos, xFace, bm)
		ySolid = model.FaceSolid(pos, yFace, bm)
		zSolid = model.FaceSolid(pos, zFace, bm)
		for _, bBBox := range bBBoxs {
			if xSolid && deltas[0] != 0{
				xOffset = pBBoxBeforeTick.XOffset(bBBox, out[0].leastOffset)
				out[0].appendOffset(xOffset, bBBox)
			}
			if ySolid && deltas[1] != 0{
				yOffset = pBBoxBeforeTick.YOffset(bBBox, out[1].leastOffset)
				out[1].appendOffset(yOffset, bBBox)
			}
			if zSolid && deltas[2] != 0{
				zOffset = pBBoxBeforeTick.ZOffset(bBBox, out[2].leastOffset)
				out[2].appendOffset(zOffset, bBBox)
			}
		}
	}
	return m.firstCollideAxis(out, deltas)
}

func (lb *leastBoxOffset) appendOffset(offset float64, bBBox cube.BBox) {
	if isCloserToZero(offset, lb.leastOffset) > 0 {
		lb.leastOffsetBBoxes = lb.leastOffsetBBoxes[:0]
		lb.leastOffset = offset
	}
	if offset == lb.leastOffset {
		lb.leastOffsetBBoxes = append(lb.leastOffsetBBoxes, bBBox)
	}
}

// bboxIntersectsSolid reports whether any block collision box overlaps pBBox.
func (m *Movement) bboxIntersectsSolid(pBBox cube.BBox) bool {
	bm := m.state.BlockMap()
	for _, pos := range m.blockPositionsInBBox(pBBox) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok {
			continue
		}
		for _, bBBox := range model.BBox(pos, bm) {
			if pBBox.IntersectsWith(bBBox) {
				return true
			}
		}
	}
	return false
}
