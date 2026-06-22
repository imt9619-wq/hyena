package movements

import (
	"fmt"
	"math"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (m *Movement) doMotions() {
	m.setSlipperiness()
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
	if mgl64.FloatEqualThreshold(m.position[1], float64(m.state.BlockMap().Dimension().Range()[0]), utils.Negligible){
		m.position[1] = float64(m.state.BlockMap().Dimension().Range()[0])
	}
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

func (m *Movement) setSlipperiness() {
	if !m.onGround {
		m.slipperiness = AirborneSlipperiness
		return
	}
	pPosUnder1b := m.position.Sub(mgl64.Vec3{0, 0.5, 0})
	bboxUnder1B := utils.TinyBBoxOnBBoxFace(utils.PlayerBBox(pPosUnder1b), cube.FaceDown)
	pCubePos := cube.PosFromVec3(pPosUnder1b)
	for pos, bl := range utils.BlockInBBox(bboxUnder1B, m.state.BlockMap()){
		if pos != pCubePos{
			continue
		}
		fmt.Printf("block under: %T\n", bl)
		switch bl.(type){
		case block.Slime:
			m.slipperiness = 0.8
		case block.PackedIce:
			m.slipperiness = 0.98
		case block.BlueIce:
			m.slipperiness = 0.989
		default:
			m.slipperiness = DefaultSlipperiness
		}
		return
	}
	m.slipperiness = DefaultSlipperiness
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
	if m.onGround {
		m.velocity[1] = JumpSpeed
	}
	m.state.Player().SetFlag(packet.InputFlagJumping)
	m.state.Player().SetFlag(packet.InputFlagJumpCurrentRaw)
}

func (m *Movement) run() {
	ps := m.state.Player()
	yawRad := float64(ps.Yaw) * (math.Pi / 180)
	sinD := math.Sin(yawRad)
	cosD := math.Cos(yawRad)

	if m.onGround {
		accel := 0.1 * SprintMovementMult * math.Pow(0.6/m.slipperiness, 3)
		m.velocity[0] += accel * sinD
		m.velocity[2] += accel * cosD
	} else {
		airAccel := 0.02 * SprintMovementMult
		m.velocity[0] += airAccel * sinD
		m.velocity[2] += airAccel * cosD
	}

	if m.isjumping && m.onGround {
		m.velocity[0] += SprintJumpBoost * sinD
		m.velocity[2] += SprintJumpBoost * cosD
	}

	m.state.Player().SetFlag(packet.InputFlagSprinting)
}
