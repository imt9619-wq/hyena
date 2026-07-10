package handler

import (
	"time"

	"github.com/imt9619-wq/hyena/game"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (c *Connection) tick() {
	defer c.handler.OnAfterTick(c)
	c.handler.OnBeforeTick(c)
	<-c.state.Exec(c.gameStateTick)
	c.Conn.Flush()
}

func (c *Connection) gameStateTick(q *game.Qx) {
	c.state.Tick()
	for pk := range c.state.FlushPackets(){
		if _, ok := pk.(*packet.PlayerAuthInput); ok && c.onUi.Load(){
			continue
		}
		c.WritePacket(pk)
	}
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
