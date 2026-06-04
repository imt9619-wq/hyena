package handler

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
)

func newPlayerState() *playerState {
	return &playerState{
		RWMutex: &sync.RWMutex{},
		velocity:       mgl32.Vec3{},
		onGround:       true,
		playerPosition: mgl32.Vec3{},
		pitch:          0,
		yaw:            0,
	}
}

type playerState struct {
	*sync.RWMutex
	playerPosition mgl32.Vec3
	pitch    	   float32
	yaw            float32
	velocity       mgl32.Vec3
	onGround       bool
}