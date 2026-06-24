package game

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft"
)

type moveBuf struct {
	lastTickInBuf  uint
	firstTickInBuf uint
	bufSize        int
	buf            []*InMovement
}

type InMovement struct {
	position  mgl64.Vec3
	velocity  mgl64.Vec3
	isrunning bool
	isjumping bool
	onGround  bool
}

func newMoveBuf(conn *minecraft.Conn) *moveBuf{
	mb := &moveBuf{}
	mb.lastTickInBuf = 0
	mb.firstTickInBuf = 0
	mb.bufSize = int(conn.GameData().PlayerMovementSettings.RewindHistorySize)
	mb.buf = make([]*InMovement, 0, mb.bufSize)
	return mb
}

func (mb *moveBuf) addTick(newMove *InMovement, tick uint){
	if len(mb.buf) == mb.bufSize{
		mb.lastTickInBuf += 1
		mb.firstTickInBuf += 1
		mb.buf[0] = nil
		mb.buf = mb.buf[1:]
		mb.buf = append(mb.buf, newMove)
		return
	}
}

func (mb *moveBuf) bufRecover(lastTick uint){
	mb.firstTickInBuf = mb.lastTickInBuf
}