package hblock

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

type Block interface {
	Slipperiness() float64
	Climbable() bool
}

func WblockToHblock(b world.Block) Block{
	switch b.(type){
	case block.BlueIce:
		return BlueIce{}
	case block.Slime:
		return Slime{}
	case block.PackedIce:
		return PackedIce{}
	case block.Ladder:
		return Ladder{}
	case block.Vines:
		return Vines{}
	default:
		return DefaultPorp{}
	}
}