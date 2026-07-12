package pathfind

import "github.com/go-gl/mathgl/mgl32"

type node struct {
	g, h, f  int
	position mgl32.Vec3
	parent   *node
}