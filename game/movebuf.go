package game

import (
	"fmt"
	"time"

	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/sandertv/gophertunnel/minecraft"
)

type move struct{
	simInMove *movements.InMovement
	simResult *movements.OutMovement
}

type moveBuf struct {
	lastTickInBuf  uint
	firstTickInBuf uint
	bufSize        int
	buf            []*move
}

func newMoveBuf(conn *minecraft.Conn) *moveBuf{
	mb := &moveBuf{}
	mb.bufSize = int(conn.GameData().PlayerMovementSettings.RewindHistorySize)
	if mb.bufSize <= 0{
		mb.bufSize = 1
	}
	mb.buf = make([]*move, 0, mb.bufSize*4)
	return mb
}

// add an outMovement to buffer after finishing a tick for movement simulation
func (mb *moveBuf) addTick(newInMove *movements.InMovement, newOutMove *movements.OutMovement){
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
	mb.buf = append(mb.buf, &move{simInMove: newInMove, simResult: newOutMove})
}

func (mb *moveBuf) outMoveWithTick(tick uint) (*move, bool){
	if mb.firstTickInBuf <= tick && mb.lastTickInBuf >= tick && len(mb.buf) != 0{
		return mb.buf[int(tick-mb.firstTickInBuf)], true
	}
	return nil, false
}

func (gs *GameState) ReSimMoveAtTick(tick uint, modF func(*movements.InMovement)){
	now := time.Now()
	startTick := tick + 1
	mb := gs.moveBuf
	out, ok := mb.outMoveWithTick(startTick)
	if !ok{
		in := gs.player.spiltInMovement(gs.in)
		modF(in)
		gs.player.copyMovement(&in.AMovement)
		gs.in = in.Input
		return
	}
	in := out.simInMove
	modF(in)
	for currTick := startTick; currTick <= mb.lastTickInBuf; currTick++{
		ind := currTick-mb.firstTickInBuf
		mb.buf[ind].simInMove = in
		out := gs.player.movement.SimMovements(in)
		in.AMovement = out.AMovement
		if currTick != mb.lastTickInBuf{
			in.Input = mb.buf[ind+1].simInMove.Input
		}
	}
	fmt.Printf("(%0.3fms)resim pos %v to %v(in: %v(tick: %v))\n", time.Since(now).Seconds()*1000, gs.player.position, in.Position, out.simInMove.Position, startTick)
	gs.player.copyMovement(&in.AMovement)
}