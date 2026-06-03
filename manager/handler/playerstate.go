package handler

import (
	"sync"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
)

func newPlayerState() *playerState {
	return &playerState{
		RWMutex: &sync.RWMutex{},
		force:          mgl32.Vec3{},
		velocity:       mgl32.Vec3{},
		onGround:       bool(true),
		playerPosition: mgl32.Vec3{},
		pitch:          0,
		yaw:            0,
		onReset:        &atomic.Bool{},
	}
}




type playerState struct {
	*sync.RWMutex
	playerPosition mgl32.Vec3
	pitch    	   float32
	yaw            float32
	force          mgl32.Vec3
	velocity       mgl32.Vec3
	onGround       bool
	onReset        *atomic.Bool
}