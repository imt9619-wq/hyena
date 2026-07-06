package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type playerState struct {
	position mgl32.Vec3
    velocity mgl32.Vec3
    onGround bool
	baseSpeed float32

	in input.Inputs
	movement *movements.Movement
}

func newPlayerState(conn *minecraft.Conn, move *movements.Movement) *playerState {
	ps := &playerState{
		position: conn.GameData().PlayerPosition,
		in: input.Inputs{},
		velocity: mgl32.Vec3{},
		onGround: false,
		baseSpeed: float32(movements.DefaultBaseSpeed),
		movement: move,
	}
	ps.in.Yaw = conn.GameData().Yaw
	ps.in.Pitch = conn.GameData().Pitch
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
	in.Input = ps.in
	in.BaseSpeed = ps.baseSpeed 
	return in
}

func (ps *playerState) setPlayerAuthInputWithPlayerState(pk *packet.PlayerAuthInput){
	pk.Pitch, pk.InteractPitch = ps.in.Pitch, ps.in.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = ps.in.Yaw, ps.in.Yaw, ps.in.Yaw
	pk.Position = ps.position
}

func (ps *playerState) copyMovement(out *movements.AMovement){
	ps.position = out.Position
	ps.velocity = out.Velocity
	ps.onGround = out.OnGround
	ps.baseSpeed = out.BaseSpeed
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