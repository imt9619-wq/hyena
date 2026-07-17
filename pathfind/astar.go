package pathfind

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

type node struct {
	g, h, f  float64
	position cube.Pos
	parent   *node
}
