package handler

import (
	"math"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
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

func (m *movement) checkCollision(){

}

func istrue(b *atomic.Bool) bool {
	return b.Load()
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
