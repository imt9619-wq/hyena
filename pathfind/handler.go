package pathfind

import (
	"fmt"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/dbgshape"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type PathFindHandler struct {
	handler.NopConnHandler
    c          *handler.Connection
    shape      *dbgshape.ShapeCache
    world      *blockmap.BlockMap
    packets    chan packet.Packet
    callerName string
	targetPos  mgl32.Vec3
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
	h.world = blockmap.NewBlockMap(c.Conn, utils.NopPacketBuffer{})
	h.c = c
	go h.startRunning()
}

func (h *PathFindHandler) startRunning() {
	for {
		select {
		case <-h.c.Closed():
			return
		case pk := <-h.packets:
			if ph, ok := packetToPacketHandler[pk.ID()]; ok{
				ph.handle(h, pk)
			}
		}
	}
}

func (h *PathFindHandler) OnDisconnect(c *handler.Connection, reason string){
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
	h.shape.Close()
}

func (h *PathFindHandler) OnPacket(ctx *handler.Context, pk packet.Packet){
	if _, ok := packetToPacketHandler[pk.ID()]; ok{
		h.packets <-pk
	}
}