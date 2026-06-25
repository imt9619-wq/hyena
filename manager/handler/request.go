package handler

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (c *Connection) requestSubChunkInQuery() {
	querySub := c.state.BlockMap().SubChunkInQuery()
	for dim, query := range querySub {
		queryl := len(query)
		if queryl == 0 {
			continue
		}
		dimen, _ := world.DimensionByID(dim)
		r := dimen.Range()
		offsets := make([]protocol.SubChunkOffset, 0, r.Height()>>4)
		var pos int32
		for cpos, chunkSub := range query{
			offsets = offsets[:0]
			for subPos := range chunkSub {
				pos = subPos
				break
			}
			for subPos := range chunkSub {
				offsets = append(offsets, [3]int8{0, int8(subPos - pos), 0})
			}
			c.WritePacket(&packet.SubChunkRequest{
				Dimension: int32(dim),
				Position:  protocol.SubChunkPos{cpos[0], pos, cpos[1]},
				Offsets:   offsets,
			})
		}
		
	}
}

func (c *Connection) requestNetworkStackLatency(pk *packet.NetworkStackLatency) {
	c.WritePacket(&packet.NetworkStackLatency{
		Timestamp:     pk.Timestamp * 1000000,
		NeedsResponse: true,
	})
}
