package handler

import "github.com/sandertv/gophertunnel/minecraft"

type ConnBuf struct {
	*minecraft.Conn
	H ConnHandler
}

func NewConnBuf(conn *minecraft.Conn) *ConnBuf{
	return &ConnBuf{
		Conn: conn,
		H: DefaultHandler{},
	}
}


func (cb *ConnBuf) Handle(h ConnHandler){
	cb.H = h
}