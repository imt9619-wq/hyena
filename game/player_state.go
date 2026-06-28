package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type playerState struct {
	Position mgl32.Vec3
    Velocity mgl32.Vec3
    OnGround bool
	baseSpeed float32

	in movements.Inputs
}

func newPlayerState(conn *minecraft.Conn) *playerState {
	ps := &playerState{
		Position: conn.GameData().PlayerPosition,
		in: movements.Inputs{},
		Velocity: mgl32.Vec3{},
		OnGround: false,
		baseSpeed: float32(movements.DefaultBaseSpeed),
	}
	ps.in.Yaw = conn.GameData().Yaw
	ps.in.Pitch = conn.GameData().Pitch
	return ps
}

func (ps *playerState) setPlayerAuthInputWithPlayerState(pk *packet.PlayerAuthInput){
	pk.Pitch, pk.InteractPitch = ps.in.Pitch, ps.in.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = ps.in.Yaw, ps.in.Yaw, ps.in.Yaw
	pk.Position = ps.Position
}

func (ps *playerState) splitInMovement(flags *protocol.Bitset) *movements.InMovement{
	in := &movements.InMovement{}
	ps.in.InputFlags = flags
	in.Position = ps.Position
	in.OnGround = ps.OnGround
	in.Velocity = ps.Velocity
	in.Input = ps.in
	in.BaseSpeed = ps.baseSpeed 
	return in
}

func (ps *playerState) copyOutMovement(out *movements.OutMovement){
	ps.Position = out.Position
	ps.Velocity = out.Velocity
	ps.OnGround = out.OnGround
	ps.baseSpeed = out.BaseSpeed
}