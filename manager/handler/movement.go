package handler

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	playerHeight       = float32(1.8)
	playerWidth        = float32(0.6)
	defaultSlipperiness = float32(0.6)
	sprintMovementMult  = float32(1.3)
	sprintJumpBoost     = float32(0.2)
	jumpSpeed           = float32(0.42)
	negligibleMomentum  = float64(0.003)
)

type movement struct {
	state     *gameState
	isrunning bool
	isjumping bool

	intersectBlocks    map[cube.Pos]struct{}
	blockPosScratch    []cube.Pos
	floorPointsScratch []float64
}

func newMovement(state *gameState) *movement {
	return &movement{
		state:           state,
		intersectBlocks: make(map[cube.Pos]struct{}, 16),
	}
}

func (m *movement) tick() {
	m.doMotions()
	m.applyVelocity()
	m.checkCollision()
}

func (m *movement) doMotions() {
	m.applyHorizontalMovement()
	if m.isjumping {
		m.jump()
	}
}

// cheakCollision will check if the player will collision with a block after apply velocity on its
// position, if collision player velocity towards the axis where the player BBox and block BBox
// collided will be set to 0, the position will also be set the position where the player BBox is
// just touching the block BBox in the axis that they collided
func (m *movement) checkCollision() {
	ps := m.state.player
	pPosBeforeTick := ps.position.Sub(ps.velocity)
	pBBoxBeforeTick := playerBBox(pPosBeforeTick)
	c := m.walkPath(pBBoxBeforeTick)
	for i := 0; i < c.nIndies; i++ {
		index := c.indies[i]
		ps.position[index] = ps.position[index] - ps.velocity[index] + c.closestBox[index].leastOffset
		if (index == 1 && ps.velocity[1] < c.closestBox[index].leastOffset && ps.velocity[1] < 0) || c.closestBox[index].leastOffset == 0 {
			ps.onGround = true
		}
		if ps.velocity[index] == c.closestBox[index].leastOffset && c.closestBox[index].leastOffset != 0 {
			ps.onGround = false
		}
		if ps.velocity[index] != c.closestBox[index].leastOffset || c.closestBox[index].leastOffset == 0 {
			ps.velocity[index] = 0
		}
	}
}

func playerBBox(pos mgl32.Vec3) cube.BBox {
	halfW := playerWidth / 2
	return cube.Box(
		float64(pos[0]-halfW),
		float64(pos[1]),
		float64(pos[2]-halfW),
		float64(pos[0]+halfW),
		float64(pos[1]+playerHeight),
		float64(pos[2]+halfW),
	)
}

func (m *movement) walkPath(pBBox cube.BBox) collideOffsets {
	var foot axisOffsets
	c := m.collideWithOffsets(pBBox, m.state.player.velocity, &foot)
	return m.walkStairs(pBBox, c)
}

type axisOffsets [3]leastBoxOffset

func (a *axisOffsets) reset(deltas mgl32.Vec3) {
	for i := 0; i < 3; i++ {
		a[i].leastOffset = deltas[i]
		a[i].leastOffsetBBoxes = a[i].leastOffsetBBoxes[:0]
	}
}

type collideOffsets struct {
	closestBox axisOffsets
	indies     [3]int
	nIndies    int
}

func collideIndiesContains(c collideOffsets, axis int) bool {
	for i := 0; i < c.nIndies; i++ {
		if c.indies[i] == axis {
			return true
		}
	}
	return false
}

func (m *movement) walkStairs(pBBox cube.BBox, c collideOffsets) collideOffsets {
	ps := m.state.player
	if !((ps.onGround && ps.velocity[1] == 0) && !collideIndiesContains(c, 1)) {
		return c
	}
	if c.nIndies == 0 {
		return c
	}
	index := c.indies[0]
	offset := c.closestBox[index].leastOffset
	if c.nIndies > 1 || ps.velocity[index] == offset {
		return c
	}
	bboxWithOffset := c.closestBox[index].leastOffsetBBoxes
	var stepHeight float64
	for _, bBBox := range bboxWithOffset {
		currentStepHeight := bBBox.Max()[1] - pBBox.Min()[1]
		if currentStepHeight > stepHeight {
			stepHeight = currentStepHeight
		}
		if currentStepHeight > 0.6 {
			return c
		}
	}
	if stepHeight == 0 {
		return c
	}

	walkStairDelta := ps.velocity
	walkStairDelta[1] = float32(stepHeight)
	var up axisOffsets
	upC := m.collideWithOffsets(pBBox, walkStairDelta, &up)
	if upC.closestBox[1].leastOffset < float32(stepHeight) {
		return c
	}

	raisedBBox := pBBox.Translate(mgl64.Vec3{0, stepHeight, 0})
	if m.bboxIntersectsSolid(raisedBBox) {
		return c
	}

	horizontalVel := ps.velocity
	horizontalVel[1] = 0
	var step axisOffsets
	stepC := m.collideWithOffsets(raisedBBox, horizontalVel, &step)
	finalBBox := raisedBBox.Translate(mgl64.Vec3{
		float64(stepC.closestBox[0].leastOffset),
		0,
		float64(stepC.closestBox[2].leastOffset),
	})
	if m.bboxIntersectsSolid(finalBBox) {
		return c
	}
	stepC.closestBox[1].leastOffset = float32(stepHeight)
	return m.firstCollideAxis(&stepC.closestBox, ps.velocity)
}

// will return the expected final delta on each axis of the player last tick position to its future tick,
// most of the time, only one offset on an axis is gonna be applied as the player will reach that axis of
// the bbox before the another two and if so the player will be stopped on that point already so the player
// wont actually be able to reach the another two axis, although there is some extreme case where offset to
// velocity radio is the same for two axis(or even three) where then the player will reach the two axis at
// the same time, therefore we are returning a slice instead of just a float for those cases(int is for
// index, float is the actual offset)
func (m *movement) firstCollideAxis(closestBox *axisOffsets, deltas mgl32.Vec3) collideOffsets {
	var radios [3]float32
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

type leastBoxOffset struct {
	leastOffset       float32
	leastOffsetBBoxes []cube.BBox
}

// return >= 0 if currentOffset is closer to 0 else < -1
func isCloserToZero(currentOffset float64, leastOffset float64) float64 {
	return math.Abs(leastOffset) - math.Abs(currentOffset)
}

func deltaIsZero(d mgl32.Vec3) bool {
	return d[0] == 0 && d[1] == 0 && d[2] == 0
}

func (m *movement) collideWithOffsets(pBBoxBeforeTick cube.BBox, deltas mgl32.Vec3, out *axisOffsets) collideOffsets {
	out.reset(deltas)
	if deltaIsZero(deltas) {
		return collideOffsets{closestBox: *out}
	}

	bm := m.state.blockMap
	xSolid, ySolid, zSolid := true, true, true
	var xOffset, yOffset, zOffset float64
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
			if xSolid {
				xOffset = pBBoxBeforeTick.XOffset(bBBox, float64(out[0].leastOffset))
				out[0].appendOffset(xOffset, bBBox)
			}
			if ySolid {
				yOffset = pBBoxBeforeTick.YOffset(bBBox, float64(out[1].leastOffset))
				out[1].appendOffset(yOffset, bBBox)
			}
			if zSolid {
				zOffset = pBBoxBeforeTick.ZOffset(bBBox, float64(out[2].leastOffset))
				out[2].appendOffset(zOffset, bBBox)
			}
		}
	}
	return m.firstCollideAxis(out, deltas)
}

func (lb *leastBoxOffset) appendOffset(offset float64, bBBox cube.BBox) {
	if isCloserToZero(offset, float64(lb.leastOffset)) > 0 {
		lb.leastOffsetBBoxes = lb.leastOffsetBBoxes[:0]
		lb.leastOffset = float32(offset)
	}
	if offset == float64(lb.leastOffset) {
		lb.leastOffsetBBoxes = append(lb.leastOffsetBBoxes, bBBox)
	}
}

func (m *movement) blockPositionsInBBox(bbox cube.BBox) []cube.Pos {
	min := bbox.Min()
	max := bbox.Max()
	m.blockPosScratch = m.blockPosScratch[:0]
	for x := int(math.Floor(min[0])); x <= int(math.Floor(max[0])); x++ {
		for y := int(math.Floor(min[1])); y <= int(math.Floor(max[1])); y++ {
			for z := int(math.Floor(min[2])); z <= int(math.Floor(max[2])); z++ {
				m.blockPosScratch = append(m.blockPosScratch, cube.Pos{x, y, z})
			}
		}
	}
	return m.blockPosScratch
}

// bboxIntersectsSolid reports whether any block collision box overlaps pBBox.
func (m *movement) bboxIntersectsSolid(pBBox cube.BBox) bool {
	bm := m.state.blockMap
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

// This function will get all the block pos that the player BBox will came accoss from the position of last
// tick to the future position after applying the velocity on the position,
func (m *movement) intersection(pBBox cube.BBox, deltas mgl32.Vec3) map[cube.Pos]struct{} {
	clear(m.intersectBlocks)
	pCorners := pBBox.Corners()
	velocity := mgl32Vec3Tomgl64Vec3(deltas)

	for _, corner := range pCorners {
		for index, val := range corner {
			if velocity[index] == 0 {
				continue
			}
			for _, point := range floorFloatBetweenAB(val, val+velocity[index], &m.floorPointsScratch) {
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
				m.intersectBlocks[Mgl64Vec3ToCubePos(newVec3)] = struct{}{}
			}
		}
	}
	return m.intersectBlocks
}

// since a block position is gonna be a 3 integer array, we just need to get the integer in range of the integer
// of the player last tick position to the future tick position, for example, if a player when from [2.5, 0, 1.1]
// to [-5.1, 2, 0.6](pretty big changes for a tick, it is rare just for demostration) then the integer x,y,z in
// range of the last tick to the future tick is -5,-4,-3,-2,-1,0,1,2 for x. 0,1,2 for y. 1 for z.
// Then we can input all the value of the axis to threeDLine() to get all the points the player BBox will come
// accross on its way to the future tick position, since it is possible to have mutiple point of the same point
// of one axis of movement, we need the changes of integer from all axis instead of just inputing plotting one point
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

func Mgl64Vec3ToCubePos(v mgl64.Vec3) cube.Pos {
	return cube.Pos{
		int(math.Floor(v[0])),
		int(math.Floor(v[1])),
		int(math.Floor(v[2])),
	}
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

func mgl32Vec3Tomgl64Vec3(v mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{float64(v[0]), float64(v[1]), float64(v[2])}
}

func (m *movement) applyVelocity() {
	ps := m.state.player
	if !ps.onGround {
		ps.velocity[1] = (ps.velocity[1] - 0.08) * 0.98
	}
	ps.position = ps.position.Add(ps.velocity)
}

func (m *movement) startRunning() {
	m.state.Exec(func(q *Qx) { m.isrunning = true })
}

func (m *movement) stopRunning() {
	m.state.Exec(func(q *Qx) { m.isrunning = false })
}

func (m *movement) startJumping() {
	m.state.Exec(func(q *Qx) { m.isjumping = true })
}

func (m *movement) stopJumping() {
	m.state.Exec(func(q *Qx) { m.isjumping = false })
}

// applyHorizontalMovement applies vanilla per-axis friction then sprint input acceleration.
// See https://www.mcpk.wiki/wiki/Horizontal_Movement_Formulas
func (m *movement) applyHorizontalMovement() {
	ps := m.state.player
	slipperiness := float32(1.0)
	if ps.onGround {
		slipperiness = defaultSlipperiness
	}
	friction := slipperiness * 0.91

	mx := ps.velocity[0] * friction
	mz := ps.velocity[2] * friction
	if math.Abs(float64(mx)) < negligibleMomentum {
		mx = 0
	}
	if math.Abs(float64(mz)) < negligibleMomentum {
		mz = 0
	}

	if !m.isrunning {
		ps.velocity[0] = mx
		ps.velocity[2] = mz
		return
	}

	yawRad := float64(ps.yaw) * (math.Pi / 180)
	sinD := float32(math.Sin(yawRad))
	cosD := float32(math.Cos(yawRad))

	if ps.onGround {
		accel := float32(0.1) * sprintMovementMult * float32(math.Pow(0.6/float64(slipperiness), 3))
		ps.velocity[0] = mx + accel*sinD
		ps.velocity[2] = mz + accel*cosD
	} else {
		airAccel := float32(0.02) * sprintMovementMult
		ps.velocity[0] = mx + airAccel*sinD
		ps.velocity[2] = mz + airAccel*cosD
	}

	if m.isjumping && ps.onGround {
		ps.velocity[0] += sprintJumpBoost * sinD
		ps.velocity[2] += sprintJumpBoost * cosD
	}
}

func (m *movement) jump() {
	ps := m.state.player
	if ps.onGround {
		ps.velocity[1] = jumpSpeed
		ps.onGround = false
	}
}

func rotationToPitchAndYaw(r mgl32.Vec3) (yaw, pitch float32) {
	xz := math.Sqrt(math.Pow(float64(r[0]), 2) + math.Pow(float64(r[2]), 2))
	mag := math.Sqrt(math.Pow(xz, 2) + math.Pow(float64(r[1]), 2))

	pitch, yaw = float32(18/math.Pi), float32(18/math.Pi)
	if xz > 0.003 {
		yaw = float32(math.Acos(float64(r[2])/xz) * 180 / math.Pi)
	}
	if mag > 0.003 {
		pitch = float32(math.Acos(xz/mag) * 180 / math.Pi)
	}
	return
}

func xzSpeed(v mgl32.Vec3) float32 {
	return float32(math.Sqrt(math.Pow(float64(v[0]), 2) + math.Pow(float64(v[2]), 2)))
}
