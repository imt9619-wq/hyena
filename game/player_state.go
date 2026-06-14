package game

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
)

type playerState struct {
	Position mgl32.Vec3
	Pitch    float32
	Yaw      float32
	Velocity mgl32.Vec3
	OnGround bool
}

func newPlayerState(conn *minecraft.Conn) *playerState {
	ps := &playerState{
		Velocity: mgl32.Vec3{},
		OnGround: true,
	}
	ps.Yaw = conn.GameData().Yaw
	ps.Position = conn.GameData().PlayerPosition
	ps.Pitch = conn.GameData().Pitch
	return ps
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

func (ps *playerState) SetVelocityTo(v mgl32.Vec3) {
	ps.Velocity = v
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