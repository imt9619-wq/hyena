package handler

import (
	"fmt"
)

type ConnHandler interface {
	HandleDisconnect(*ConnBuf, string)
}

type DefaultHandler struct{}

func (h DefaultHandler) HandleDisconnect(cb *ConnBuf, reason string) {
	fmt.Printf("Disconnected: %s", reason)
}
