package handler

import (
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/event"
	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/game"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Connection wraps a Minecraft protocol connection with game state and a user handler.
type Connection struct {
	*minecraft.Conn
    handler    Handler
    closeOnce  *sync.Once
    closed     chan struct{}
    state      *game.GameState
    onUi       *atomic.Bool
    entInWorld *EntInWorld
}

func NewConnection(conn *minecraft.Conn, h Handler) *Connection {
	c := &Connection{
		Conn:      conn,
		handler:   h,
		closed:    make(chan struct{}),
		closeOnce: &sync.Once{},
		onUi: &atomic.Bool{},
		entInWorld: newEntInWorld(conn),
	}
	c.state = game.NewGameState(conn)
	go c.startTicking()
	return c
}

func (c *Connection) SetHandler(h Handler) {
	c.handler = h
}

func (c *Connection) Handler() Handler {
	return c.handler
}

func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.state.Close()
		c.Conn.Close()
	})
}

func (c *Connection) Closed() chan struct{}{
	return c.closed
}

func (c *Connection) HandlePacket(pk packet.Packet){
	ctx := event.C(c)
	if c.handler.OnPacket(ctx, pk); ctx.Cancelled() {
		return
	}
	switch pk := pk.(type) {
	case *packet.NetworkStackLatency:
		c.handleNetworkStackLatency(pk)
	case *packet.LevelChunk:
		c.handleLevelChunk(pk)
	case *packet.NetworkChunkPublisherUpdate:
		c.handleNetworkChunkPublisherUpdate(pk)
	case *packet.ChunkRadiusUpdated:
		c.handleChunkRadiusUpdated(pk)
	case *packet.UpdateAttributes:
		c.handleUpdateAttributes(pk)
	case *packet.SetActorMotion:
		c.handleSetActorMotion(pk)
	case *packet.UpdateBlock:
		c.handleUpdateBlock(pk)
	case *packet.SubChunk:
		c.handleSubChunk(pk)
	case *packet.MovePlayer:
		c.handleMovePlayer(pk)
	case *packet.CorrectPlayerMovePrediction:
		c.handleCorrectPlayerMovePrediction(pk)
	case *packet.InventoryContent:
		c.handleInventoryContent(pk)
	case *packet.MobEquipment:
		c.handleMobEquipment(pk)
	case *packet.ModalFormRequest:
		c.handleModalFormRequest(pk)
	case *packet.InventorySlot:
		c.handleInventorySlot(pk)
	case *packet.PlayerList:
		c.handlePlayerList(pk)
	case *packet.UpdateSubChunkBlocks:
		c.handleUpdateSubChunkBlocks(pk)
	}
}

func (c *Connection) GameState() *game.GameState{
	return c.state
}

func (c *Connection) requestNetworkStackLatency(pk *packet.NetworkStackLatency) {
	c.WritePacket(&packet.NetworkStackLatency{
		Timestamp:     pk.Timestamp * 1000000,
		NeedsResponse: true,
	})
}

func (c *Connection) ExcuteCommand(cmd string){
	pk := &packet.CommandRequest{
		CommandLine: cmd,
		CommandOrigin: protocol.CommandOrigin{
			Origin: protocol.CommandOriginPlayer,
			UUID: uuid.New(),
		},
	}
	c.WritePacket(pk)
}