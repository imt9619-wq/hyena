package sim

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
)

// Player holds local player kinematics for simulation.
type Player struct {
	Position mgl32.Vec3
	Pitch    float32
	Yaw      float32
	Velocity mgl32.Vec3
	OnGround bool
}

// NewPlayer initialises player state from a spawned connection.
func NewPlayer(conn *minecraft.Conn) *Player {
	ps := &Player{
		Velocity: mgl32.Vec3{},
		OnGround: true,
	}
	ps.Yaw = conn.GameData().Yaw
	ps.Position = conn.GameData().PlayerPosition
	ps.Pitch = conn.GameData().Pitch
	return ps
}

func (p *Player) sinNCosOfSpeed() (sinD, cosD float32) {
	speed := xzSpeed(p.Velocity)
	xVel := p.Velocity[0]
	zVel := p.Velocity[2]

	sinD = 0
	cosD = 1
	if speed > 0.003 {
		sinD = xVel / speed
		cosD = zVel / speed
	}
	return
}

// SetSpeedTo preserves direction and sets horizontal speed magnitude.
func (p *Player) SetSpeedTo(s float32) {
	sinD, cosD := p.sinNCosOfSpeed()
	p.Velocity[0] = s * sinD
	p.Velocity[2] = s * cosD
}

func xzSpeed(v mgl32.Vec3) float32 {
	return float32(math.Sqrt(math.Pow(float64(v[0]), 2) + math.Pow(float64(v[2]), 2)))
}
