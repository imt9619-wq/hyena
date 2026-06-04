package handler

import (
	"sync"

	"github.com/sandertv/gophertunnel/minecraft"
)

type ConnBuf struct {
	*minecraft.Conn
	h ConnHandler
	movements *playerMovement
	closeOnce *sync.Once
	closed chan struct{}
	sc *sessionConf
}


func NewConnBuf(conn *minecraft.Conn, h ConnHandler) *ConnBuf {
	cb := &ConnBuf{
		Conn: conn,
		h: h,
		closed: make(chan struct{}),
		closeOnce: &sync.Once{},
	}
	cb.sc = NewsessionConf(conn)
	cb.movements = newPlayerMovement(cb.sc)

	cb.startTicking()
	return cb
}

func (cb *ConnBuf) StartRunning(){
	cb.movements.startRunning()
}

func (cb *ConnBuf) StopRunning(){
	cb.movements.stopRunning()
}

func (cb *ConnBuf) Handle(h ConnHandler){
	cb.h = h
}

func (cb *ConnBuf) H() ConnHandler{
	return cb.h
}


func (cb *ConnBuf) Close(){
	cb.closeOnce.Do(func ()  {
		close(cb.closed)
		cb.Conn.Close()
	})
}