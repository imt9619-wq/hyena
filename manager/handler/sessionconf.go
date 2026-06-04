package handler

import (
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// this struct contains minecraft game information of player in the world
type sessionConf struct {
	conn *minecraft.Conn
	entityRuntimeID uint64

	playerState *playerState

	flushedTick *atomic.Int32
	packetQueue []packet.Packet
}


func (sc *sessionConf) flush(){
	defer sc.playerState.RUnlock()
	defer sc.flushedTick.Add(1)
	sc.playerState.RLock()
	//sc.writePlayerAuthInput()
}


func (sc *sessionConf) writePlayerAuthInput(){
	ps := sc.playerState
	sc.conn.WritePacket(&packet.PlayerAuthInput{
		Pitch: ps.pitch,
		Yaw: ps.yaw,
		Position: ps.playerPosition,
		MoveVector: mgl32.Vec2([]float32{ps.velocity[0], ps.velocity[2]}),

	})
}


func NewsessionConf(conn *minecraft.Conn) *sessionConf {
	sc := &sessionConf{
		conn: conn,
		entityRuntimeID: 0,
		playerState: newPlayerState(conn),
		flushedTick: &atomic.Int32{},
		packetQueue: make([]packet.Packet, 0, 10),
	}
	sc.entityRuntimeID = conn.GameData().EntityRuntimeID
	
	return sc
}
