package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (m *Movement) doMotions() {
	m.setSlipperiness()
	m.setOnClimb()
	m.applyHorizontalMovement()
	if m.isjumping {
		m.jump()
	}
	if m.isrunning {
		m.run()
	}
	m.applyGravity()
}

func (m *Movement) setOnClimb(){
	if m.world.Hblock(cube.PosFromVec3(m.position)).Climbable(){
		m.onClimb = true
	}else{
		m.onClimb = false
	}
}

func (m *Movement) applyGravity() {
	if !m.onGround && !m.onClimb {
		m.velocity[1] = (m.velocity[1] - 0.08) * 0.98
		return
	}
	if m.onClimb && !m.isjumping{
		m.velocity[1] = ClimbSpeed * -1
	}
}

func (m *Movement) setSlipperiness() {
	if !m.onGround {
		m.slipperiness = AirborneSlipperiness
		return
	}
	bl := m.world.Hblock(cube.PosFromVec3(m.position.Sub(mgl64.Vec3{0, 0.5, 0})))
	m.slipperiness = bl.Slipperiness()
}

// applyHorizontalMovement applies vanilla per-axis friction then sprint input acceleration.
// See https://www.mcpk.wiki/wiki/Horizontal_Movement_Formulas
func (m *Movement) applyHorizontalMovement() {
	friction := m.slipperiness * SlipperinessToFriction
	mx := m.velocity[0] * friction
	mz := m.velocity[2] * friction
	if math.Abs(mx) < MomentumThreshold {
		mx = 0
	}
	if math.Abs(mz) < MomentumThreshold {
		mz = 0
	}
	m.velocity[0] = mx
	m.velocity[2] = mz
}

func (m *Movement) movementMultiplier() float64{
	return 0.98
}

func (m *Movement) jump() {
	if m.onClimb{
		m.velocity[1] = ClimbSpeed
		m.setFlag(packet.InputFlagWantUp)
		return
	}
	if m.onGround {
		m.velocity[1] = JumpSpeed
	}
	m.setFlag(packet.InputFlagJumping)
	m.setFlag(packet.InputFlagJumpCurrentRaw)
}

func (m *Movement) run() {
	yawRad := m.yaw * (math.Pi / 180)
	sinD := math.Sin(yawRad)
	cosD := math.Cos(yawRad)

	if m.onGround {
		accel := m.baseSpeed * SprintMovementMult * m.movementMultiplier() * math.Pow(0.6/m.slipperiness, 3)
		m.velocity[0] += accel * sinD
		m.velocity[2] += accel * cosD
	} else {
		airAccel := 0.02 * SprintMovementMult
		m.velocity[0] += airAccel * sinD
		m.velocity[2] += airAccel * cosD
	}

	if m.isjumping && m.onGround && !m.onClimb{
		m.velocity[0] += SprintJumpBoost * sinD
		m.velocity[2] += SprintJumpBoost * cosD
	}

	m.setFlag(packet.InputFlagSprinting)
	m.setFlag(packet.InputFlagUp)
}
