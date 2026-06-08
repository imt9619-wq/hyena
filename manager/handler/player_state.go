package handler

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
)

type playerState struct {
	position mgl32.Vec3
	pitch    float32
	yaw      float32
	velocity mgl32.Vec3
	onGround bool
}

func newPlayerState(conn *minecraft.Conn) *playerState {
	ps := &playerState{
		velocity: mgl32.Vec3{},
		onGround: true,
	}
	ps.yaw = conn.GameData().Yaw
	ps.position = conn.GameData().PlayerPosition
	ps.pitch = conn.GameData().Pitch
	return ps
}

func (ps *playerState) sinNCosOfSpeed() (sinD, cosD float32) {
	speed := xzSpeed(ps.velocity)
	xVel := ps.velocity[0]
	zVel := ps.velocity[2]

	sinD = float32(0)
	cosD = float32(1)
	if speed > 0.003 {
		sinD = xVel / speed
		cosD = zVel / speed
	}
	return
}

func (ps *playerState) setSpeedTo(s float32) {
	sinD, cosD := ps.sinNCosOfSpeed()
	ps.velocity[0] = s*sinD
	ps.velocity[2] = s*cosD
}