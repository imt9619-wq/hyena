package handler

import (
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/input"
)

func (c *Connection) StartRunning(once bool) {
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.W.Pressed = true
			i.Sprint.Pressed = true
			i.Sprint.PressOnce = once
		})
	})
}

func (c *Connection) StopRunning() {
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.W = input.KeyPress{}
			i.Sprint = input.KeyPress{}
		})
	})
}

func (c *Connection) StartJumping(once bool) {
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.Space.Pressed = true
			i.Space.PressOnce = once
		})		
	})
}

func (c *Connection) StopJumping() {
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.Space = input.KeyPress{}
		})		
	})
}

func (c *Connection) SetYaw(yaw float32){
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.Yaw = yaw
		})
	})
}

func (c *Connection) StartSneaking(once bool){
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.Shift.Pressed = true
			i.Shift.PressOnce = once
		})
	})
}

func (c *Connection) StopSneaking() {
	c.state.Exec(func(q *game.Qx) {
		q.SetInput(func(i *input.Inputs) {
			i.Shift = input.KeyPress{}
		})
	})
}