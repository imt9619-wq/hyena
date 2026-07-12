package utils

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type BlockSourse interface {
	world.BlockSource
	BlockModel(cube.Pos, uint8) (world.BlockModel, bool)
}

type PacketBuffer interface{
	Append(packet.Packet)
}