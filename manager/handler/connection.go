package handler

import (
	"sync"

	"github.com/sandertv/gophertunnel/minecraft"
)

// Connection wraps a Minecraft protocol connection with game state and a user handler.
type Connection struct {
	*minecraft.Conn
	handler   Handler
	movement  *movement
	closeOnce *sync.Once
	closed    chan struct{}
	state     *gameState
}

func NewConnection(conn *minecraft.Conn, h Handler) *Connection {
	c := &Connection{
		Conn:      conn,
		handler:   h,
		closed:    make(chan struct{}),
		closeOnce: &sync.Once{},
	}
	c.state = newGameState(conn)
	c.movement = newMovement(c.state)
	c.startTicking()
	c.state.startRunningQueue(c)
	return c
}

func (c *Connection) StartRunning() {
	c.movement.startRunning()
}

func (c *Connection) StopRunning() {
	c.movement.stopRunning()
}

func (c *Connection) StartJumping() {
	c.movement.startJumping()
}

func (c *Connection) StopJumping() {
	c.movement.stopJumping()
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
		c.Conn.Close()
	})
}
