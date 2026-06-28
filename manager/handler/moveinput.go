package handler

import (
	"github.com/imt9619-wq/hyena/game"
)

func (c *Connection) StartRunning(once bool) {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().W.Pressed = true
		c.state.Inputs().Sprint.Pressed = once
	})
}

func (c *Connection) StopRunning() {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().W.Pressed = false
		c.state.Inputs().Sprint.Pressed = false
	})
}

func (c *Connection) StartJumping(once bool) {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().Space.Pressed = true
		c.state.Inputs().Space.PressOnce = once
	})
}

func (c *Connection) StopJumping() {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().Space.Pressed = false
	})
}

func (c *Connection) SetYaw(yaw float32){
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().Yaw = yaw
	})
}