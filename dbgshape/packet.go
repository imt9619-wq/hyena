package dbgshape

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	Delete uint8 = iota
	Add
)

type DebugShape struct {
	Opts uint8 // Operation
	// ShapeID is needed when deleting shapes, we should have new id for each shape unless deleting
	ShapeID uuid.UUID
	Shape   AShape
}

func init() {
	packet.RegisterPacketFromClient((&DebugShape{}).ID(), func() packet.Packet {
		return &DebugShape{}
	})
}

func (ds *DebugShape) ID() uint32 {
	return 23
}

func (pk *DebugShape) Marshal(io protocol.IO) {
	io.Uint8(&pk.Opts)
	io.UUID(&pk.ShapeID)
	io.Uint8(&pk.Shape.Shape)
	io.Vec3(&pk.Shape.StartPoint)
	io.Vec3(&pk.Shape.EndPoint)
}
