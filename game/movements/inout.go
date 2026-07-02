package movements

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

type AMovement struct {
	Position   mgl32.Vec3
    Velocity   mgl32.Vec3
	
	BaseSpeed  float32
    OnGround   bool
    
	Input Inputs
}

type InMovement AMovement

func (m *Movement) copyInMovement(in *InMovement) {
	m.velocity = utils.Mgl32Vec3Tomgl64Vec3(in.Velocity.Add(in.Input.ServerSpeedAdd))
	m.position = utils.Mgl32Vec3Tomgl64Vec3(in.Position).Sub(mgl64.Vec3{0, utils.NetworkOffset, 0})
	m.position = utils.RoundVecTo5Decimal(m.position)
	m.yaw = float64(in.Input.Yaw)
	m.onGround = in.OnGround
	m.Inputs = in.Input
	m.baseSpeed = float64(in.BaseSpeed)
}

type OutMovement AMovement

func (m *Movement) splitOutMovement() *OutMovement{
	out := &OutMovement{}
	out.Velocity = utils.Mgl64Vec3Tomgl32Vec3(m.velocity)
	out.Position = utils.Mgl64Vec3Tomgl32Vec3(m.position.Add(mgl64.Vec3{0, utils.NetworkOffset, 0}))
	out.OnGround = m.onGround 
	out.Input = m.Inputs
	out.BaseSpeed = float32(m.baseSpeed)
	m.flag = nil
	return out
}

// doesnt copy input state and addedSpeed
func (out *AMovement) CopyOutToMove(move *AMovement){
	move.Position = out.Position
	move.Velocity = out.Velocity
	move.OnGround = out.OnGround
	move.BaseSpeed = out.BaseSpeed
}

func (provide *AMovement) CopyInputToMove(receiver *AMovement){
	receiver.Input = provide.Input
}