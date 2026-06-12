package handler

import "time"

func (c *Connection) tick() {
	<-c.state.Exec(c.gameStateTick)
}

func (c *Connection) gameStateTick(q *Qx) {
	defer c.state.flush()
	c.state.session.Tick()
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
}
