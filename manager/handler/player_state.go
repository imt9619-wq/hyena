package handler

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
)

type playerState struct {
	*sync.RWMutex
	position mgl32.Vec3
	pitch    float32
	yaw      float32
	velocity mgl32.Vec3
	onGround bool
}

func newPlayerState(conn *minecraft.Conn) *playerState {
	ps := &playerState{
		RWMutex:  &sync.RWMutex{},
		velocity: mgl32.Vec3{},
		onGround: true,
	}
	ps.Lock()
	defer ps.Unlock()
	ps.yaw = conn.GameData().Yaw
	ps.position = conn.GameData().PlayerPosition
	ps.pitch = conn.GameData().Pitch
	return ps
}
