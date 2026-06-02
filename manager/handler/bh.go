package handler

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"




func (cb *ConnBuf) BhNSL(pk *packet.NetworkStackLatency){
	if !pk.NeedsResponse{
		return
	}
	cb.WritePacket(&packet.NetworkStackLatency{
		Timestamp: pk.Timestamp*1000000,
		NeedsResponse: pk.NeedsResponse,
	})
}