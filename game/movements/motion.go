package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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
	if m.Shift.Pressed{
		m.sneak()
	}
}

func (m *Movement) sneak(){
	m.setFlag(packet.InputFlagSneaking)
	m.setFlag(packet.InputFlagStartSneaking)
	m.setFlag(packet.InputFlagSneakDown)
	m.setFlag(packet.InputFlagPersistSneak)
	m.setFlag(packet.InputFlagSneakCurrentRaw)
	m.setFlag(packet.InputFlagSneakToggleDown)
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
		m.setFlag(packet.InputFlagWantDown)
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
		m.setFlag(packet.InputFlagWantUp)
		return
	}
	if m.onGround {
		m.velocity[1] = JumpSpeed
		m.setFlag(packet.InputFlagStartJumping)
	}
	m.setFlag(packet.InputFlagJumping)
	m.setFlag(packet.InputFlagJumpCurrentRaw)
}

func (m *Movement) run() {
	yawRad := m.yaw * (math.Pi / 180)
	// sin is reverse for minecraft yaw
	sinF := -math.Sin(yawRad)
	cosF := math.Cos(yawRad)
	dirRad := (m.keyOffsets() + m.yaw) * (math.Pi / 180)
	sinD := -math.Sin(dirRad)
	cosD := math.Cos(dirRad)

	if m.onGround {
		accel := m.baseSpeed * m.movementMultiplier() * math.Pow(0.6/m.slipperiness, 3)
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
	m.setHorizontalFlags()
}

func (m *Movement) setHorizontalFlags(){
	if m.IsSprinting(){
		m.setFlag(packet.InputFlagSprintDown)
		m.setFlag(packet.InputFlagSprinting)
		m.setFlag(packet.InputFlagStartSprinting)
	}
	if m.IsUpWalk(){
		m.setFlag(packet.InputFlagUp)
	}
	if m.IsDownWalk(){
		m.setFlag(packet.InputFlagDown)
	}
	if m.IsRightWalk(){
		m.setFlag(packet.InputFlagRight)
	}
	if m.IsLeftWalk(){
		m.setFlag(packet.InputFlagLeft)
	}
	
	switch m.keyOffsets(){
	case 45:
		m.setFlag(packet.InputFlagUpRight)
	case 135:
		m.setFlag(packet.InputFlagDownRight)
	case -135:
		m.setFlag(packet.InputFlagDownLeft)
	case -45:
		m.setFlag(packet.InputFlagUpLeft)
	default:
	} 
}
