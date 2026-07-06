package movements

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/utils"
)

type AMovement struct {
	Position mgl32.Vec3
    Velocity mgl32.Vec3

    JumpCooldown int
    BaseSpeed    float32
    OnGround     bool
}

type InMovement struct{
	AMovement
	Input input.Inputs
} 

func (m *Movement) copyInMovement(in *InMovement) {
	m.velocity = utils.Mgl32Vec3Tomgl64Vec3(in.Velocity.Add(in.Input.ServerSpeedAdd))
	m.position = utils.Mgl32Vec3Tomgl64Vec3(in.Position).Sub(mgl64.Vec3{0, utils.NetworkOffset, 0})
	m.position = utils.RoundVecTo5Decimal(m.position)
	m.yaw = float64(in.Input.Yaw)
	m.onGround = in.OnGround
	m.Inputs = in.Input
	m.baseSpeed = float64(in.BaseSpeed)
	m.jumpCooldown = in.JumpCooldown
	m.flag = MovementFlags{}
}

type OutMovement struct{
	AMovement
	Flag MovementFlags
}

func (m *Movement) splitOutMovement() *OutMovement{
	out := &OutMovement{}
	out.Velocity = utils.Mgl64Vec3Tomgl32Vec3(m.velocity)
	out.Position = utils.Mgl64Vec3Tomgl32Vec3(m.position.Add(mgl64.Vec3{0, utils.NetworkOffset, 0}))
	out.OnGround = m.onGround 
	out.BaseSpeed = float32(m.baseSpeed)
	out.Flag = m.flag
	out.JumpCooldown = m.jumpCooldown
	return out
}
