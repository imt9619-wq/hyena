package game

import (
	"iter"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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

func (mb *moveBuf) iterFromTick(startTick uint) iter.Seq2[uint, *movements.OutMovement]{
	return func(yield func(uint, *movements.OutMovement) bool) {
		if mb.bufSize == 0 || len(mb.buf) == 0{
			return
		}
		if mb.firstTickInBuf <= startTick && mb.lastTickInBuf >= startTick && len(mb.buf) != 0{
			for ind := startTick; ind <= mb.lastTickInBuf; ind++{
				if !yield(ind-mb.firstTickInBuf, mb.buf[ind-mb.firstTickInBuf]){
					return 
				}
			}
			return 
		}
	}
}

func (gs *GameState) ReSimMovements(pk *packet.CorrectPlayerMovePrediction){
	mb := gs.moveBuf
	startTick := uint(pk.Tick)
	currInMove := &movements.InMovement{}
	yaw, _:= utils.RotationToPitchAndYaw(mgl32.Vec3{pk.Rotation[0], 0, pk.Rotation[1]}) 
	_, ok := mb.outMoveWithTick(gs.tick)
	if !ok || gs.tick == startTick{
		gs.player.Position = pk.Position
		gs.player.OnGround = pk.OnGround
		gs.player.Yaw = yaw
		return
	}
	var newOut *movements.OutMovement
	for currTick := startTick; currTick < mb.lastTickInBuf; currTick++{
		ind := currTick-mb.firstTickInBuf
		currOut := mb.buf[ind]
		currOut.Position = pk.Position
		currOut.OnGround = pk.OnGround
		currOut.Yaw = yaw
		currOut.CopyOutToIn(currInMove)
		newOut = gs.movement.SimMovement(currInMove)
		mb.buf[ind+1] = newOut
	}
	lastTickOut := mb.buf[mb.lastTickInBuf-mb.firstTickInBuf]
	gs.copyOutMovement(lastTickOut)
}