package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft"
)

type playerState struct {
	position     mgl32.Vec3
    velocity     mgl32.Vec3
    onGround     bool
    baseSpeed    float32
    jumpCooldown int

    movement *movements.Movement
}

func newPlayerState(conn *minecraft.Conn, move *movements.Movement) *playerState {
	ps := &playerState{
		position: conn.GameData().PlayerPosition,
		velocity: mgl32.Vec3{},
		onGround: false,
		baseSpeed: float32(movements.DefaultBaseSpeed),
		movement: move,
	}
	return ps
}

func (ps *playerState) doMove(in *movements.InMovement) *movements.OutMovement{
	out := ps.movement.SimMovements(in)
	ps.copyMovement(&out.AMovement)
	return out
}

func (ps *playerState) spiltInMovement(input input.Inputs) *movements.InMovement{
	in := &movements.InMovement{}
	in.Position = ps.position
	in.OnGround = ps.onGround
	in.Velocity = ps.velocity
	in.Input = input
	in.BaseSpeed = ps.baseSpeed 
	in.JumpCooldown = ps.jumpCooldown
	return in
}

func (ps *playerState) copyMovement(out *movements.AMovement){
	ps.position = out.Position
	ps.velocity = out.Velocity
	ps.onGround = out.OnGround
	ps.baseSpeed = out.BaseSpeed
	ps.jumpCooldown = out.JumpCooldown
}

func (ps *playerState) Position() mgl32.Vec3{
	return ps.position
}

func (ps *playerState) Velocity() mgl32.Vec3{
	return ps.velocity
}

func (ps *playerState) OnGround() bool{
	return ps.onGround
}