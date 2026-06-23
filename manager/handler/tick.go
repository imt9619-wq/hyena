package handler

import (
	"time"

	"github.com/imt9619-wq/hyena/game"
)

func (c *Connection) tick() {
	<-c.state.Exec(c.gameStateTick)
}

func (c *Connection) gameStateTick(q *game.Qx) {
	c.state.Tick()
	c.movement.Tick()
	c.requestSubChunkInQuery()
	c.WritePacket(c.state.PlayerAuthInputWithState())
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
