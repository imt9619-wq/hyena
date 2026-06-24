package movements

import "github.com/go-gl/mathgl/mgl64"

type InMovement struct {
	position  mgl64.Vec3
	velocity  mgl64.Vec3
	isrunning bool
	isjumping bool
}