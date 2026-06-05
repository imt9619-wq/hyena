package handler

import "time"

func (c *Connection) tick() {
	defer c.state.flush()
	c.movement.tick()
}

func (c *Connection) startTicking() {
	ticker := time.NewTicker(50 * time.Millisecond)

	go func() {
		for {
			select {
			case <-c.closed:
				ticker.Stop()
				return
			case <-ticker.C:
				c.tick()
			}
		}
	}()
	c.tick()
}
