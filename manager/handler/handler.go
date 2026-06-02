package handler

import (
	"fmt"
)

type ConnHandler interface {
	HandleDisconnect(*ConnBuf, string)
	HandleJoin(*ConnBuf)
}

type DefaultHandler struct{}

func (h DefaultHandler) HandleDisconnect(cb *ConnBuf, reason string) {
	fmt.Printf("Disconnected: %s\n", reason)
}

func (h DefaultHandler) HandleJoin(cb *ConnBuf){
	fmt.Printf("%s has joined the server: %s\n", cb.IdentityData().DisplayName, cb.RemoteAddr())
}