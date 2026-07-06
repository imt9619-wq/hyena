package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

func (m *Movement) doMotions() {
	m.setBBoxFunc()
	m.setOnClimb()
	m.setSlipperiness()
	m.applyHorizontalMovement()
	if m.Space.Pressed {
		m.jump()
	}
	if !m.IsStop() {
		m.run()
	}
	m.applyGravity()
}

func (m *Movement) setBBoxFunc(){
	if m.Shift.Pressed{
		m.bboxFunc = utils.PlayerSneakBBox
	}else{
		m.bboxFunc = utils.PlayerBBox
	}
}

func (m *Movement) setOnClimb(){
	if m.world.Hblock(cube.PosFromVec3(m.position)).Climbable(){
		m.onClimb = true
		m.flag.OnClimb = true
	}else{
		m.onClimb = false
	}
}

func (m *Movement) applyGravity() {
	if !m.onGround && !m.onClimb {
		m.velocity[1] = (m.velocity[1] - 0.08) * 0.98
		return
	}
	if m.onClimb && !m.Space.Pressed && !m.Shift.Pressed{
		m.velocity[1] = ClimbSpeed * -1
		m.flag.WantDown = true
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

func (m *Movement) jump() {
	if m.onClimb{
		m.velocity[1] = ClimbSpeed
		m.flag.WantUp = true
		return
	}
	if m.onGround {
		m.velocity[1] = JumpSpeed
		m.flag.StartedJumping = true
	}
}

func (m *Movement) run() {
	yawRad := m.yaw * (math.Pi / 180)
	// sin is reverse for minecraft yaw
	sinF := -math.Sin(yawRad)
	cosF := math.Cos(yawRad)
	dirRad := (m.KeyOffsets() + m.yaw) * (math.Pi / 180)
	sinD := -math.Sin(dirRad)
	cosD := math.Cos(dirRad)

	if m.onGround {
		accel := m.baseSpeed * m.MovementMultiplier() * math.Pow(0.6/m.slipperiness, 3)
		m.velocity[0] += accel * sinD
		m.velocity[2] += accel * cosD

		if m.Space.Pressed && !m.onClimb && m.IsSprinting(){
			m.velocity[0] += SprintJumpBoost * sinF
			m.velocity[2] += SprintJumpBoost * cosF
		}
	} else {
		m.velocity[0] += AirborneAccelration * sinD
		m.velocity[2] += AirborneAccelration * cosD
	}
}
