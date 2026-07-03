package handler

import (
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/movements"
)

func (c *Connection) StartRunning(once bool) {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().W.Pressed = true
		c.state.Inputs().Sprint.Pressed = true
		c.state.Inputs().Sprint.PressOnce = once
	})
}

func (c *Connection) StopRunning() {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().W = movements.KeyPress{}
		c.state.Inputs().Sprint = movements.KeyPress{}
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
		c.state.Inputs().Space = movements.KeyPress{}
	})
}

func (c *Connection) SetYaw(yaw float32){
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().Yaw = yaw
	})
}

func (c *Connection) StartSneaking(once bool){
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().Shift.Pressed = true
		c.state.Inputs().Shift.PressOnce = once
	})
}

func (c *Connection) StopSneaking() {
	c.state.Exec(func(q *game.Qx) {
		c.state.Inputs().Shift = movements.KeyPress{}
	})
}