package handler

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"


func (cb *ConnBuf) BhDisconnect(reason string){
	cb.h.HandleDisconnect(cb, reason)
}

func (cb *ConnBuf) BhNetworkStackLatency(pk *packet.NetworkStackLatency){
	if !pk.NeedsResponse{
		return
	}
	cb.WritePacket(&packet.NetworkStackLatency{
		Timestamp: pk.Timestamp*1000000,
		NeedsResponse: pk.NeedsResponse,
	})
}

func (cb *ConnBuf) BhMoveActorAbsolute(pk *packet.MoveActorAbsolute){
	
}


func (cb *ConnBuf) BhStartGame(pk *packet.StartGame){
	cb.sc.entityRuntimeID = pk.EntityRuntimeID
}

func (cb *ConnBuf) BhJoin(){
	cb.h.HandleJoin(cb)
}