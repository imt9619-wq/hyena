package handler

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func (c *Connection) requestSubChunks(pk *packet.LevelChunk) {
	// TODO
}

func (c *Connection) requestNetworkStackLatency(pk *packet.NetworkStackLatency) {
	c.WritePacket(&packet.NetworkStackLatency{
		Timestamp:     pk.Timestamp * 1000000,
		NeedsResponse: true,
	})
}