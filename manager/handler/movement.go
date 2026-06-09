package handler

import (
	"math"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

type movement struct {
	state          *gameState
	isrunning *atomic.Bool
	isjumping *atomic.Bool
	isleftwalk *atomic.Bool
	isrightwalk *atomic.Bool
	isbackwalk *atomic.Bool
	iswalk *atomic.Bool
	inputFlags map[int]struct{}
	lastTickInputFlags map[int]struct{}
}

func newMovement(state *gameState) *movement {
	m := &movement{
		state:         state,
		isrunning: &atomic.Bool{},
		isjumping: &atomic.Bool{},
		isleftwalk: &atomic.Bool{},
		isrightwalk: &atomic.Bool{},
		isbackwalk: &atomic.Bool{},
		iswalk: &atomic.Bool{},
		inputFlags: make(map[int]struct{}, 5),
		lastTickInputFlags: make(map[int]struct{}, 5),
	}
	m.isrunning.Store(false)
	m.isjumping.Store(false)
	m.isleftwalk.Store(false)
	m.isrightwalk.Store(false)
	m.isbackwalk.Store(false)
	m.iswalk.Store(false)
	return m
}

func (m *movement) tick() {
	m.doMotions()
	m.applyVelocity()
	// m.checkCollision()
}

// checkCollision will get all new BBox the player has position into, 
// we assume there is no collision in last tick so we can check less blocks BBox, 
// we will check what block pos is touched by a 3d object where it is shaped by mapping 
// all the points the pBBox has intercepted when moving from the last tick position to the 
// current tick position, it can have 0 points(if no move) to 14 points. if a collision happens 
// on a given axis of the player BBox, the speed that is toward that axis will be reset to 0, 
// and the player position will be set the point where the player BBox collied axis is the 
// touching the collied block axis
func (m *movement) checkCollision(){
	ps := m.state.player
	pheight := float32(1.8)
	pwidth := float32(0.6)
	pPosBeforeTick := ps.position.Sub(ps.velocity)
	pBBoxBeforeTick := cube.Box(float64(pPosBeforeTick[0]), 
								float64(pPosBeforeTick[1]), 
								float64(pPosBeforeTick[2]), 
								float64(pPosBeforeTick[0]+pwidth), 
								float64(pPosBeforeTick[1]+pheight), 
								float64(pPosBeforeTick[2]+pwidth),
							)
	_ = m.intersection(pBBoxBeforeTick)
}

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
// will return false if that point is impossible to reach
func (m *movement) threeDLine(i mgl64.Vec3, direction mgl64.Vec3, inputPointIndex int, inputPointValue float64) (mgl64.Vec2, bool) {
	var outputPointsPair mgl64.Vec2
	if direction[inputPointIndex] == 0{	
		return outputPointsPair, false
	}
	iToIndexOffset := (i[inputPointIndex] - inputPointValue) / direction[inputPointIndex]

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

func istrue(b *atomic.Bool) bool {
	return b.Load()
}

func mgl32Vec3Tomgl64Vec3(m mgl32.Vec3) mgl64.Vec3 {
	return mgl64.Vec3([]float64{float64(m[0]), float64(m[1]), float64(m[2])})
}

func (m *movement) doMotions(){
	if istrue(m.isrunning) {m.run()}
	if istrue(m.isjumping) {m.jump()}
}

func (m *movement) applyVelocity() {
	gravity := float32(-0.08)
	drag := float32(0.98)
	ps := m.state.player
	if !ps.onGround{
		ps.velocity[1] = (ps.velocity[1] + gravity) * drag
	}
	m.state.player.position.Add(m.state.player.velocity)
}

func (m *movement) startRunning() {
	m.isrunning.Store(true)
}

func (m *movement) stopRunning() {
	m.isrunning.Store(false)
}

func (m *movement) startJumping() {
	m.isjumping.Store(true)
}

func (m *movement) stopJumping() {
	m.isjumping.Store(false)
}

func (m *movement) run() {
	slipperiness := float32(0.6)
	movementMult := float32(1.3)
	effectsMult := float32(1)
	ps := m.state.player

	jumpBoost := float32(0.2)
	if !istrue(m.isjumping) {
		jumpBoost = 0
	}

	yawRad := float64(ps.yaw) * (math.Pi / 180)
	speed := xzSpeed(ps.velocity)

	momentum := speed * slipperiness * 0.91
	acceleration := float32(0.1) * movementMult * effectsMult * float32(math.Pow(0.6/float64(slipperiness), 3))
	if !ps.onGround {
		acceleration = 0
	}

	sinD, cosD := ps.sinNCosOfSpeed()
	ps.velocity[0] = momentum + acceleration*sinD + jumpBoost*float32(math.Sin(yawRad))
	ps.velocity[2] = momentum + acceleration*cosD + jumpBoost*float32(math.Cos(yawRad))
}

func (m *movement) jump() {
	ps := m.state.player
	jumpSpeed := float32(0.42)
	
	if ps.onGround {
		ps.position[1] = jumpSpeed
	}
}

func rotationToPitchAndYaw(r mgl32.Vec3) (yaw, pitch float32) {
	xz := math.Sqrt(math.Pow(float64(r[0]), 2) + math.Pow(float64(r[2]), 2))
	mag := math.Cbrt(math.Pow(xz, 2) + math.Pow(float64(r[1]), 2))

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
