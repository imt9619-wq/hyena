package handler

import (
	"sync"

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
}

func NewConnection(conn *minecraft.Conn, h Handler) *Connection {
	c := &Connection{
		Conn:      conn,
		handler:   h,
		closed:    make(chan struct{}),
		closeOnce: &sync.Once{},
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
	default:
	}
}

func (c *Connection) GameState() *game.GameState{
	return c.state
}