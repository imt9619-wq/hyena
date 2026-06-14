package movements

import (
	"math"

	"github.com/imt9619-wq/hyena/game"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (m *Movement) doMotions() {
	m.applyHorizontalMovement()
	if m.isjumping {
		m.jump()
	}
	if m.isrunning {
		m.run()
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
	if m.isrunning {
		return
	}
	m.state.Exec(func(q *game.Qx) {
		m.isrunning = true
		m.state.Player().SetFlag(packet.InputFlagStartSprinting)
	})
}

func (m *Movement) StopRunning() {
	if !m.isrunning {
		return
	}
	m.state.Exec(func(q *game.Qx) {
		m.isrunning = false
		m.state.Player().SetFlag(packet.InputFlagStopSprinting)
	})
}

func (m *Movement) StartJumping() {
	if m.isjumping {
		return
	}
	m.state.Exec(func(q *game.Qx) {
		m.isjumping = true
		m.state.Player().SetFlag(packet.InputFlagJumpPressedRaw)
		m.state.Player().SetFlag(packet.InputFlagJumpCurrentRaw)
		m.state.Player().SetFlag(packet.InputFlagStartJumping)
	})
}

func (m *Movement) StopJumping() {
	if !m.isjumping {
		return
	}
	m.state.Exec(func(q *game.Qx) {
		m.isjumping = false
		m.state.Player().SetFlag(packet.InputFlagJumpReleasedRaw)
	})
}

func (m *Movement) slipperiness() float64 {
	slipperiness := float64(1.0)
	if m.onGround {
		slipperiness = defaultSlipperiness
	}
	return slipperiness
}

func (m *Movement) friction() float64 {
	slipperiness := m.slipperiness()
	friction := slipperiness * 0.91
	return friction
}

// applyHorizontalMovement applies vanilla per-axis friction then sprint input acceleration.
// See https://www.mcpk.wiki/wiki/Horizontal_Movement_Formulas
func (m *Movement) applyHorizontalMovement() {
	friction := m.friction()
	mx := m.velocity[0] * friction
	mz := m.velocity[2] * friction
	if math.Abs(mx) < momentumThreshold {
		mx = 0
	}
	if math.Abs(mz) < momentumThreshold {
		mz = 0
	}
	m.velocity[0] = mx
	m.velocity[2] = mz
}

func (m *Movement) jump() {
	if m.onGround {
		m.velocity[1] = jumpSpeed
	}
	m.state.Player().SetFlag(packet.InputFlagJumping)
	m.state.Player().SetFlag(packet.InputFlagJumpCurrentRaw)
}

func (m *Movement) run() {
	slipperiness := m.slipperiness()
	ps := m.state.Player()
	yawRad := float64(ps.Yaw) * (math.Pi / 180)
	sinD := math.Sin(yawRad)
	cosD := math.Cos(yawRad)

	if m.onGround {
		accel := 0.1 * sprintMovementMult * math.Pow(0.6/slipperiness, 3)
		m.velocity[0] += accel * sinD
		m.velocity[2] += accel * cosD
	} else {
		airAccel := 0.02 * sprintMovementMult
		m.velocity[0] += airAccel * sinD
		m.velocity[2] += airAccel * cosD
	}

	if m.isjumping && m.onGround {
		m.velocity[0] += sprintJumpBoost * sinD
		m.velocity[2] += sprintJumpBoost * cosD
	}

	m.state.Player().SetFlag(packet.InputFlagSprinting)
}
