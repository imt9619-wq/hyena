package movements

import (
	"math"

	"github.com/imt9619-wq/hyena/game"
)

func (m *Movement) doMotions() {
	m.applyHorizontalMovement()
	if m.isjumping {
		m.jump()
	}
	m.applyGravity()
}

func (m *Movement) applyGravity() {
	if !m.onGround {
		m.velocity[1] = (m.velocity[1] - 0.08) * 0.98
	}
}

func (m *Movement) applyVelocity() {
	m.position = m.position.Add(m.velocity)
}

func (m *Movement) StartRunning() {
	m.state.Exec(func(q *game.Qx) { m.isrunning = true })
}

func (m *Movement) StopRunning() {
	m.state.Exec(func(q *game.Qx) { m.isrunning = false })
}

func (m *Movement) StartJumping() {
	m.state.Exec(func(q *game.Qx) { m.isjumping = true })
}

func (m *Movement) StopJumping() {
	m.state.Exec(func(q *game.Qx) { m.isjumping = false })
}

// applyHorizontalMovement applies vanilla per-axis friction then sprint input acceleration.
// See https://www.mcpk.wiki/wiki/Horizontal_Movement_Formulas
func (m *Movement) applyHorizontalMovement() {
	ps := m.state.Player()
	slipperiness := float64(1.0)
	if m.onGround{
		slipperiness = defaultSlipperiness
	}
	friction := slipperiness * 0.91

	mx := m.velocity[0] * friction
	mz := m.velocity[2] * friction
	if math.Abs(mx) < momentumThreshold {
		mx = 0
	}
	if math.Abs(mz) < momentumThreshold {
		mz = 0
	}

	if !m.isrunning {
		m.velocity[0] = mx
		m.velocity[2] = mz
		return
	}

	yawRad := float64(ps.Yaw) * (math.Pi / 180)
	sinD := math.Sin(yawRad)
	cosD := math.Cos(yawRad)

	if m.onGround{
		accel := 0.1 * sprintMovementMult * math.Pow(0.6/slipperiness, 3)
		m.velocity[0] = mx + accel*sinD
		m.velocity[2] = mz + accel*cosD
	} else {
		airAccel := 0.02 * sprintMovementMult
		m.velocity[0] = mx + airAccel*sinD
		m.velocity[2] = mz + airAccel*cosD
	}

	if m.isjumping && m.onGround{
		m.velocity[0] += sprintJumpBoost * sinD
		m.velocity[2] += sprintJumpBoost * cosD
	}
}

func (m *Movement) jump() {
	if m.onGround{
		m.velocity[1] = jumpSpeed
	}
}

