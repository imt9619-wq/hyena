package hblock

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

type DefaultPorp struct {
	world.Block
}

func (dp DefaultPorp) Slipperiness() float64{
	return 0.6
}

func (dp DefaultPorp) Climbable() bool{
	return false
}

type BlueIce struct {
	block.BlueIce
	DefaultPorp
}

func (bi BlueIce) Slipperiness() float64{
	return bi.Friction()
}

type PackedIce struct{
	block.PackedIce
	DefaultPorp
}

func (pi PackedIce) Slipperiness() float64{
	return pi.Friction()
}

type Slime struct{
	block.Slime
	DefaultPorp
}

func (s Slime) Slipperiness() float64{
	return s.Friction()
}

type Ladder struct{
	block.Ladder
	DefaultPorp
}

func (l Ladder) Climbable() bool{
	return true
}

type Vines struct{
	block.Vines
	DefaultPorp
}

func (v Vines) Climbable() bool{
	return true
}