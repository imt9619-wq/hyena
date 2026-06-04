package handler

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)


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


/*func (cb *ConnBuf) BhMoveActorAbsolute(pk *packet.MoveActorAbsolute){
	if cb.sc.entityRuntimeID != pk.EntityRuntimeID{
		return
	}
	yaw, pitch := rotationToPitchAndYaw(pk.Rotation)
	ps := cb.sc.playerState
	ps.Lock()
	defer ps.Unlock()

	ps.playerPosition = pk.Position
	ps.velocity = mgl32.Vec3([]float32{0, 0, 0})
	ps.pitch = pitch
	ps.yaw = yaw
}*/


func (cb *ConnBuf) BhJoin(){
	cb.h.HandleJoin(cb)
}