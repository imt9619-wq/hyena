package pathfind

import (
	"fmt"
	"strings"

	"github.com/imt9619-wq/hyena/dbgshape"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type PathFindHandler struct {
	handler.NopConnHandler
    c           *handler.Connection
    shape       *dbgshape.ShapeCache
    world       *blockmap.BlockMap
    packets     chan packet.Packet
    callerName  string
    executor    *pathRunExecutor
}

func NewPathFindHandler(callerName string) *PathFindHandler{
	h := &PathFindHandler{
		shape: dbgshape.NewShapeCache(),
		packets: make(chan packet.Packet, 1024),
		callerName: strings.ToLower(callerName),
	}
	return h 
}

func (h *PathFindHandler) OnJoin(c *handler.Connection){
	fmt.Printf("%s has joined the server: %s\n", c.IdentityData().DisplayName, c.RemoteAddr())
	h.world = c.GameState().BlockMap()
	h.c = c
	h.setPathRunExector()
	go h.startPathFindRunner()
}

func (h *PathFindHandler) OnPacket(ctx *handler.Context, pk packet.Packet){
	if _, ok := packetToPacketHandler[pk.ID()]; ok{
		h.packets <-pk
	}
}

func (h *PathFindHandler) OnDisconnect(c *handler.Connection, reason string){
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
}

func (h *PathFindHandler) OnAfterTick(c *handler.Connection){
	h.executor.syncState(c.GameState().Player().SplitAMovement())
}