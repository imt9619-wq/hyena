package sim

import (
	"math"
)

type Movement struct {
	session   *Session
	isRunning bool
	isJumping bool
}

func newMovement(session *Session) *Movement {
	return &Movement{session: session}
}

func (m *Movement) tick() {
	m.doMotions()
	m.applyVelocity()
	m.checkCollision()
}

func (m *Movement) doMotions() {
	if m.isRunning {
		m.run()
	}
	if m.isJumping {
		m.jump()
	}
}

func (m *Movement) applyVelocity() {
	gravity := float32(-0.08)
	drag := float32(0.98)
	p := m.session.Player
	if !p.OnGround {
		p.Velocity[1] = (p.Velocity[1] + gravity) * drag
	}
	p.Position = p.Position.Add(p.Velocity)
}

func (m *Movement) run() {
	slipperiness := float32(0.6)
	movementMult := float32(1.3)
	effectsMult := float32(1)
	p := m.session.Player

	jumpBoost := float32(0.2)
	if !m.isJumping {
		jumpBoost = 0
	}

	yawRad := float64(p.Yaw) * (math.Pi / 180)
	speed := xzSpeed(p.Velocity)

	momentum := speed * slipperiness * 0.91
	acceleration := float32(0.1) * movementMult * effectsMult * float32(math.Pow(0.6/float64(slipperiness), 3))
	newSpeed := momentum + acceleration
	if !p.OnGround {
		acceleration = 0
	}

	sinD, cosD := p.sinNCosOfSpeed()
	p.Velocity[0] = newSpeed*sinD + jumpBoost*float32(math.Sin(yawRad))
	p.Velocity[2] = newSpeed*cosD + jumpBoost*float32(math.Cos(yawRad))
}

func (m *Movement) jump() {
	p := m.session.Player
	if p.OnGround {
		p.Velocity[1] = 0.42
	}
}
