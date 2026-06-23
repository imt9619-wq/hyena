package game

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type playerState struct {
	Position mgl32.Vec3
	Pitch    float32
	Yaw      float32
	Velocity mgl32.Vec3
	OnGround bool
	
	lastTickPos mgl32.Vec3
}

func newPlayerState(conn *minecraft.Conn) *playerState {
	ps := &playerState{
		Position: conn.GameData().PlayerPosition,
		Pitch: conn.GameData().Pitch,
		Yaw: conn.GameData().Yaw,
		Velocity: mgl32.Vec3{},
		OnGround: true,
		lastTickPos: conn.GameData().PlayerPosition,
	}
	return ps
}

func (ps *playerState) tick() {
	ps.lastTickPos = ps.Position
}

func (ps *playerState) setPlayerAuthInputWithPlayerState(pk *packet.PlayerAuthInput){
	pk.Pitch, pk.InteractYaw = ps.Pitch, ps.Pitch
	pk.Yaw, pk.InteractYaw, pk.HeadYaw = ps.Yaw, ps.Yaw, ps.Yaw
	pk.Position = ps.Position
	pk.MoveVector = mgl32.Vec2{floatSign(ps.Velocity[0]), floatSign(ps.Velocity[2])}
	pk.RawMoveVector = mgl32.Vec2{floatSign(ps.Velocity[0]), floatSign(ps.Velocity[2])}
	pk.Delta = ps.Position.Sub(ps.lastTickPos)
}

func floatSign(f float32) float32 {
	if f > 0 {
		return 1
	} else if f == 0 {
		return 0
	} else {
		return -1
	}
}

func (ps *playerState) sinNCosOfSpeed() (sinD, cosD float32) {
	speed := xzSpeed(ps.Velocity)
	xVel := ps.Velocity[0]
	zVel := ps.Velocity[2]

	sinD = float32(0)
	cosD = float32(1)
	if speed > 0.003 {
		sinD = xVel / speed
		cosD = zVel / speed
	}
	return
}

func (ps *playerState) SetSpeedTo(s float32) {
	sinD, cosD := ps.sinNCosOfSpeed()
	ps.Velocity[0] = s*sinD
	ps.Velocity[2] = s*cosD
}

func (gs *GameState) Player() *playerState {
	return gs.player
}

func xzSpeed(v mgl32.Vec3) float32 {
	return float32(math.Sqrt(math.Pow(float64(v[0]), 2) + math.Pow(float64(v[2]), 2)))
}