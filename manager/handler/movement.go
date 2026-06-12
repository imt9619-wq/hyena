package handler

import (
	"math"
	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

type movement struct {
	state     *gameState
	isrunning bool
	isjumping bool
}

func newMovement(state *gameState) *movement {
	m := &movement{
		state:     state,
		isrunning: false,
		isjumping: false,
	}
	return m
}

func (m *movement) tick() {
	m.doMotions()
	m.applyVelocity()
	m.checkCollision()
}

// cheakCollision will check if the player will collision with a block after apply velocity on its
// position, if collision player velocity towards the axis where the player BBox and block BBox
// collided will be set to 0, the position will also be set the position where the player BBox is
// just touching the block BBox in the axis that they collided
func (m *movement) checkCollision() {
	ps := m.state.player
	pheight := float32(1.8)
	pwidth := float32(0.6)
	pPosBeforeTick := ps.position.Sub(ps.velocity)
	pBBoxBeforeTick := cube.Box(float64(pPosBeforeTick[0]-pwidth/2), 
								float64(pPosBeforeTick[1]), 
								float64(pPosBeforeTick[2]-pwidth/2),
								float64(pPosBeforeTick[0]+pwidth/2), 
								float64(pPosBeforeTick[1]+pheight), 
								float64(pPosBeforeTick[2]+pwidth/2),
							)
	indies, offsets := m.collideOffsets(pBBoxBeforeTick, ps.velocity)
	m.walkStairs(pBBoxBeforeTick)
	for i, index := range indies{
		ps.position[index] = ps.position[index] - ps.velocity[index] + offsets[i]
		if (index == 1 && ps.velocity[1] < offsets[i] && ps.velocity[1] < 0) || offsets[i] == 0 {
			ps.onGround = true
		}
		if ps.velocity[index] == offsets[i] && offsets[i] != 0 {
			ps.onGround = false
		}
		if ps.velocity[index] != offsets[i] || offsets[i] == 0 {
			ps.velocity[index] = 0
		}
	}
}

func (m *movement) walkStairs(pBBox cube.BBox, indies []int, offsets []float32, lb *leastBoxOffset) {
	ps := m.state.player
	if !(ps.onGround && ps.velocity[1] == 0) || slices.Contains(indies, 1) {
		return 
	}
	index := indies[0]
	offset := offsets[0]
	if len(indies) > 1 || ps.velocity[index] == offset {
		return
	}
	bboxWithOffset := lb.leastZOffsetBBoxes
	if index == 0{
		bboxWithOffset = lb.leastXOffsetBBoxes
	}
	var stepHeight float64 = 0
	for _, bBBox := range bboxWithOffset {
		currentStepHeight := bBBox.Max()[1] - pBBox.Min()[1]
		if currentStepHeight > stepHeight{
			stepHeight = currentStepHeight
		}
		if currentStepHeight > 0.6 {
			return 
		}
	}
	walkStairDelta := ps.velocity
	walkStairDelta[1] = float32(stepHeight)
	newIndies, newOffsets := m.collideOffsets(pBBox, walkStairDelta)
	for i, index := range newIndies{
		if index == 1 {
			if newOffsets[i] < float32(stepHeight){
				return 
			}
			slices.Delete(newIndies, i, i+1)
			slices.Delete(newOffsets, i, i+1)
		}
	}
	ps.position[1] += float32(stepHeight)
}

// will return the expected final delta on each axis of the player last tick position to its future tick, 
// most of the time, only one offset on an axis is gonna be applied as the player will reach that axis of 
// the bbox before the another two and if so the player will be stopped on that point already so the player 
// wont actually be able to reach the another two axis, although there is some extreme case where offset to 
// velocity radio is the same for two axis(or even three) where then the player will reach the two axis at
// the same time, therefore we are returning a slice instead of just a float for those cases(int is for 
// index, float is the actual offset)
func (m *movement) firstCollideAxis(lb *leastBoxOffset, deltas mgl32.Vec3) ([]int, []float32) {
	var xToVRadio, yToVRadio, zToVRadio float32 = 1, 1, 1 
	if deltas[0] != 0{
		xToVRadio = float32(lb.leastXOffset)/deltas[0]
	}
	if deltas[1] != 0{
		yToVRadio = float32(lb.leastYOffset)/deltas[1]
	}
	if deltas[2] != 0{
		zToVRadio = float32(lb.leastZOffset)/deltas[2]
	}
	minOffsetRadioToVeloCity := min(xToVRadio, yToVRadio, zToVRadio)
	minOffsetIndies := make([]int, 0, 2)
	minOffsets := make([]float32, 0, 2)
	if xToVRadio == minOffsetRadioToVeloCity{
		minOffsetIndies = append(minOffsetIndies, 0)
		minOffsets = append(minOffsets, float32(lb.leastXOffset))
	}
	if yToVRadio == minOffsetRadioToVeloCity{
		minOffsetIndies = append(minOffsetIndies, 1)
		minOffsets = append(minOffsets, float32(lb.leastYOffset))
	}
	if zToVRadio == minOffsetRadioToVeloCity{
		minOffsetIndies = append(minOffsetIndies, 2)
		minOffsets = append(minOffsets, float32(lb.leastZOffset))
	}
	return minOffsetIndies, minOffsets
}

type leastBoxOffset struct {
	leastXOffset float64
	leastXOffsetBBoxes []cube.BBox
	leastYOffset float64
	leastYOffsetBBoxes []cube.BBox
	leastZOffset float64
	leastZOffsetBBoxes []cube.BBox
}

func newLeastBoxOffset(deltas mgl32.Vec3) *leastBoxOffset {
	closestBox := &leastBoxOffset{
		leastXOffset: float64(deltas[0]),
		leastYOffset: float64(deltas[1]),
		leastZOffset: float64(deltas[2]),
	} 
	closestBox.leastXOffsetBBoxes = make([]cube.BBox, 0, 3)
	closestBox.leastYOffsetBBoxes = make([]cube.BBox, 0, 3)
	closestBox.leastZOffsetBBoxes = make([]cube.BBox, 0, 3)
	return closestBox
}

func different(x float64, y float64) float64 {
	return math.Abs(x-y)
}

func (m *movement) collideOffsets(pBBoxBeforeTick cube.BBox, deltas mgl32.Vec3)  ([]int, []float32) {
	bm := m.state.blockMap
	closestBox := newLeastBoxOffset(deltas)
	xSolid, ySolid, zSolid := true, true, true
	var xOffset, yOffset, zOffset float64 = 0, 0, 0
	var xFace, yFace, zFace = cube.FaceWest, cube.FaceDown, cube.FaceNorth
	if deltas[0] > 0 {
		xFace = cube.FaceEast
	}
	if deltas[1] > 0 {
		yFace = cube.FaceUp
	}
	if deltas[2] > 0 {
		zFace = cube.FaceSouth
	}
	for pos := range m.intersection(pBBoxBeforeTick) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok{
			continue
		} 
		bBBoxs := model.BBox(pos, bm)
		xSolid =  model.FaceSolid(pos, xFace, bm)
		ySolid =  model.FaceSolid(pos, yFace, bm)
		zSolid =  model.FaceSolid(pos, zFace, bm)
		for _, bBBox := range bBBoxs{
			if xSolid {
				xOffset = pBBoxBeforeTick.XOffset(bBBox, closestBox.leastXOffset)
				if different(xOffset, closestBox.leastXOffset) > 0 {
					clear(closestBox.leastXOffsetBBoxes)
					closestBox.leastXOffset = xOffset
				}
				if xOffset == closestBox.leastXOffset{
					closestBox.leastXOffsetBBoxes = append(closestBox.leastXOffsetBBoxes, bBBox)
				}
			}
			if ySolid {
				yOffset = pBBoxBeforeTick.YOffset(bBBox, closestBox.leastYOffset)
				if different(yOffset, closestBox.leastYOffset) > 0 {
					clear(closestBox.leastYOffsetBBoxes)
					closestBox.leastYOffset = yOffset
				}
				if yOffset == closestBox.leastYOffset{
					closestBox.leastYOffsetBBoxes = append(closestBox.leastYOffsetBBoxes, bBBox)
				}
			}
			if zSolid {
				zOffset = pBBoxBeforeTick.ZOffset(bBBox, closestBox.leastZOffset)
				if different(zOffset, closestBox.leastZOffset) > 0 {
					clear(closestBox.leastZOffsetBBoxes)
					closestBox.leastZOffset = zOffset
				}
				if zOffset == closestBox.leastZOffset{
					closestBox.leastZOffsetBBoxes = append(closestBox.leastZOffsetBBoxes, bBBox)
				}
			}
		}
	}
	return m.firstCollideAxis(closestBox, deltas)
}

// This function will get all the block pos that the player BBox will came accoss from the position of last 
// tick to the future position after applying the velocity on the position, 
func (m *movement) intersection(pBBox cube.BBox) map[cube.Pos]struct{} {
	ps := m.state.player
	pCorners := pBBox.Corners()
	velocity := mgl32Vec3Tomgl64Vec3(ps.velocity)

	intersectedBlocks := make(map[cube.Pos]struct{}, 10)
	for _, corner := range pCorners {
		for index, val := range corner {
			for _, point := range floorFloatBetweenAB(val, val+velocity[index]) {
				other2Point, exist := m.threeDLine(corner, velocity, index, point)
				if !exist{
					break
				}
				var newVec3 mgl64.Vec3
				currentOther2PointIndex := 0
				for i:=0; i<=2; i++{
					if i != index{
						newVec3[i] = other2Point[currentOther2PointIndex]
						currentOther2PointIndex++
						continue
					}
					newVec3[i] = point
				}
				intersectedBlocks[Mgl64Vec3ToCubePos(newVec3)] = struct{}{}
			}
		}
	}

	return intersectedBlocks
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
func floorFloatBetweenAB(a float64, b float64) []float64 {
	if a > b {
		temp := b
		b = a
		a = temp
	}
	ceilDistance := int(math.Ceil(b-a))
	pointsInAB := make([]float64, 0, ceilDistance)
	
	for i := math.Floor(a); i <= b; i++{
		pointsInAB = append(pointsInAB, i)
	}
	return pointsInAB
}

func Mgl64Vec3ToCubePos(m mgl64.Vec3) cube.Pos {
	x := int(math.Floor(m[0]))
	y := int(math.Floor(m[1]))
	z := int(math.Floor(m[2]))
	return cube.Pos([]int{x, y, z})
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
func (m *movement) threeDLine(i mgl64.Vec3, direction mgl64.Vec3, inputPointIndex int, inputPointValue float64) (mgl64.Vec2, bool) {
	var outputPointsPair mgl64.Vec2
	if direction[inputPointIndex] == 0{	
		return outputPointsPair, false
	}
	iToIndexOffset := (inputPointValue - i[inputPointIndex]) / direction[inputPointIndex]

	nextIndex := 0
	for index, val := range i{
		if index == inputPointIndex{
			continue
		}
		outputPointsPair[nextIndex] = val + iToIndexOffset * direction[index]
		nextIndex++
	}
	return outputPointsPair, true
}

func mgl32Vec3Tomgl64Vec3(m mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3([]float64{float64(m[0]), float64(m[1]), float64(m[2])})
}

func (m *movement) doMotions(){
	if m.isrunning {m.run()}
	if m.isjumping {m.jump()}
}

func (m *movement) applyVelocity() {
	gravity := float32(-0.08)
	drag := float32(0.98)
	ps := m.state.player
	if !ps.onGround{
		ps.velocity[1] = (ps.velocity[1] + gravity) * drag
	}
	ps.position = ps.position.Add(ps.velocity)
}

func (m *movement) startRunning() {
	m.state.Exec(func(q *Qx) {m.isrunning = true})
}

func (m *movement) stopRunning() {
	m.state.Exec(func(q *Qx) {m.isrunning = false})
}

func (m *movement) startJumping() {
	m.state.Exec(func(q *Qx) {m.isjumping = true})
}

func (m *movement) stopJumping() {
	m.state.Exec(func(q *Qx) {m.isjumping = false})
}

func (m *movement) run() {
	slipperiness := float32(0.6)
	movementMult := float32(1.3)
	effectsMult := float32(1)
	ps := m.state.player

	jumpBoost := float32(0.2)
	if !m.isjumping {
		jumpBoost = 0
	}

	yawRad := float64(ps.yaw) * (math.Pi / 180)
	speed := xzSpeed(ps.velocity)

	momentum := speed * slipperiness * 0.91
	acceleration := float32(0.1) * movementMult * effectsMult * float32(math.Pow(0.6/float64(slipperiness), 3))
	newSpeed := momentum + acceleration
	if !ps.onGround {
		acceleration = 0
	}

	sinD, cosD := ps.sinNCosOfSpeed()
	ps.velocity[0] = (newSpeed)*sinD + jumpBoost*float32(math.Sin(yawRad))
	ps.velocity[2] = (newSpeed)*cosD + jumpBoost*float32(math.Cos(yawRad))
}

func (m *movement) jump() {
	ps := m.state.player
	jumpSpeed := float32(0.42)

	if ps.onGround {
		ps.velocity[1] = jumpSpeed
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
