package handler

import (
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/event"
	"github.com/imt9619-wq/hyena/game"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Connection wraps a Minecraft protocol connection with game state and a user handler.
type Connection struct {
	*minecraft.Conn
	handler   Handler
	closeOnce *sync.Once
	closed    chan struct{}
	state     *game.GameState
	onForm    *atomic.Bool
}

func NewConnection(conn *minecraft.Conn, h Handler) *Connection {
	c := &Connection{
		Conn:      conn,
		handler:   h,
		closed:    make(chan struct{}),
		closeOnce: &sync.Once{},
		onForm: &atomic.Bool{},
	}
	c.state = game.NewGameState(conn)
	c.startTicking()
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
		c.replyNetworkStackLatency(pk)
	case *packet.LevelChunk:
		c.replyLevelChunk(pk)
	case *packet.NetworkChunkPublisherUpdate:
		c.replyNetworkChunkPublisherUpdate(pk)
	case *packet.ChunkRadiusUpdated:
		c.replyChunkRadiusUpdated(pk)
	case *packet.UpdateAttributes:
		c.replyUpdateAttributes(pk)
	case *packet.SetActorMotion:
		c.replySetActorMotion(pk)
	case *packet.UpdateBlock:
		c.replyUpdateBlock(pk)
	case *packet.SubChunk:
		c.replySubChunk(pk)
	case *packet.MovePlayer:
		c.replyMovePlayer(pk)
	case *packet.CorrectPlayerMovePrediction:
		c.replyCorrectPlayerMovePrediction(pk)
	case *packet.InventoryContent:
		c.replyInventoryContent(pk)
	case *packet.MobEquipment:
		c.replyMobEquipment(pk)
	case *packet.ModalFormRequest:
		c.replyModalFormRequest(pk)
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