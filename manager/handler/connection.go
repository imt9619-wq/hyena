package handler

import (
	"sync"

	"github.com/sandertv/gophertunnel/minecraft"
)

// Connection wraps a Minecraft protocol connection with game state and a user handler.
type Connection struct {
	*minecraft.Conn
	handler   Handler
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
	c.startTicking()
	c.state.startRunningQueue(c)
	return c
}

func (c *Connection) StartRunning() {
	c.state.Exec(func(q *Qx) { c.state.session.SetRunning(true) })
}

func (c *Connection) StopRunning() {
	c.state.Exec(func(q *Qx) { c.state.session.SetRunning(false) })
}

func (c *Connection) StartJumping() {
	c.state.Exec(func(q *Qx) { c.state.session.SetJumping(true) })
}

func (c *Connection) StopJumping() {
	c.state.Exec(func(q *Qx) { c.state.session.SetJumping(false) })
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
