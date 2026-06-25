package game

import (
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft"
)

type moveBuf struct {
	lastTickInBuf  uint
	firstTickInBuf uint
	bufSize        int
	buf            []*movements.OutMovement
}

func newMoveBuf(conn *minecraft.Conn) *moveBuf{
	mb := &moveBuf{}
	mb.bufSize = int(conn.GameData().PlayerMovementSettings.RewindHistorySize)
	if mb.bufSize != 0{
		mb.buf = make([]*movements.OutMovement, 0, mb.bufSize)
	}
	return mb
}

// add an outMovement to buffer after finishing a tick for movement simulation
func (mb *moveBuf) addTick(newOutMove *movements.OutMovement){
	if mb.bufSize == 0{
		return
	}
	if len(mb.buf) == 0{
		mb.firstTickInBuf += 1
	}else if len(mb.buf) == mb.bufSize{	
		mb.firstTickInBuf += 1
		mb.buf[0] = nil
		mb.buf = mb.buf[1:]
	}
	mb.lastTickInBuf += 1
	mb.buf = append(mb.buf, newOutMove)
}

func (mb *moveBuf) outMoveWithTick(tick uint) (*movements.OutMovement, bool){
	if mb.bufSize == 0{
		return nil, false
	}
	if mb.firstTickInBuf <= tick && mb.lastTickInBuf >= tick && len(mb.buf) != 0{
		return mb.buf[int(tick-mb.firstTickInBuf)], true
	}
	return nil, false
}

func (gs *GameState) ReSimMovements(startTick uint, reSimMove *movements.AMovement){
	mb := gs.moveBuf
	currInMove := &movements.InMovement{}
	_, ok := mb.outMoveWithTick(startTick)
	if !ok || gs.tick == startTick{
		gs.player.Position = reSimMove.Position
		gs.player.Velocity = reSimMove.Velocity
		gs.player.OnGround = reSimMove.OnGround
		gs.player.Yaw = reSimMove.Yaw
		return
	}
	currOut := mb.buf[startTick-mb.firstTickInBuf]
	currOut.Position = reSimMove.Position
	currOut.Velocity = reSimMove.Velocity
	currOut.OnGround = reSimMove.OnGround
	currOut.Yaw = reSimMove.Yaw
	for currTick := startTick; currTick < mb.lastTickInBuf; currTick++{
		ind := currTick-mb.firstTickInBuf
		currOut := mb.buf[ind]
		currOut.CopyOutToIn(currInMove)
		mb.buf[ind+1] = gs.movement.SimMovement(currInMove)
	}
	lastTickOut := mb.buf[mb.lastTickInBuf-mb.firstTickInBuf]
	gs.copyOutMovement(lastTickOut)
}