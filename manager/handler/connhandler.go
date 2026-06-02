package handler

import "github.com/sandertv/gophertunnel/minecraft"

type ConnBuf struct {
	*minecraft.Conn
	H ConnHandler
}


func (cb *ConnBuf) Handle(h ConnHandler){
	cb.H = h
}