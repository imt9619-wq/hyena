package movements

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

type aMovement struct {
	Position  mgl32.Vec3
	Velocity  mgl32.Vec3
	Yaw       float32
	Pitch     float32
	OnGround  bool
	Isrunning bool
	Isjumping bool
}

type InMovement aMovement

func (m *Movement) copyInMovement(in *InMovement) {
	m.velocity = utils.Mgl32Vec3Tomgl64Vec3(in.Velocity)
	m.position = utils.Mgl32Vec3Tomgl64Vec3(in.Position).Sub(mgl64.Vec3{0, utils.NetworkOffset, 0})
	m.position = utils.RoundVecTo5Decimal(m.position)
	m.yaw = float64(in.Yaw)
	m.onGround = in.OnGround
	m.isjumping = in.Isjumping
	m.isrunning = in.Isrunning
}

type OutMovement aMovement

func (m *Movement) splitOutMovement() *OutMovement{
	out := &OutMovement{}
	out.Velocity = utils.Mgl64Vec3Tomgl32Vec3(m.velocity)
	out.Position = utils.Mgl64Vec3Tomgl32Vec3(m.position.Add(mgl64.Vec3{0, utils.NetworkOffset, 0}))
	out.OnGround = m.onGround 
	out.Yaw = float32(m.yaw)
	out.Isjumping = m.isjumping
	out.Isrunning = m.isrunning
	return out
}

func (out *OutMovement) CopyOutToIn(in *InMovement){
	in.Isjumping = out.Isjumping
	in.Isrunning = out.Isrunning
	in.Position = out.Position
	in.Velocity = out.Velocity
	in.Yaw = out.Yaw
	in.OnGround = out.OnGround
	in.Pitch = out.Pitch
}