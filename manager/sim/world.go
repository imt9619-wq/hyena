package sim

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// BlockWorld provides block collision data for movement simulation.
type BlockWorld interface {
	BlockModel(pos cube.Pos, layer uint8) (world.BlockModel, bool)
}
