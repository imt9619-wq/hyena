package handler

import (
	"github.com/imt9619-wq/hyena/manager/handler/form"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Handler receives connection lifecycle events for a server session.
type Handler interface {
	OnDisconnect(*Connection, string)
	OnJoin(*Connection)
	OnBeforeTick(*Connection)
	OnAfterTick(*Connection)
	OnPacket(*Context, packet.Packet)
	OnForm(*Context, form.Form)
}

type NopConnHandler struct{}

var _ Handler = NopConnHandler{}

func (h NopConnHandler) OnForm(*Context, form.Form){}
func (h NopConnHandler) OnPacket(*Context, packet.Packet){}
func (h NopConnHandler) OnDisconnect(*Connection, string){}
func (h NopConnHandler) OnJoin(*Connection){}
func (h NopConnHandler) OnBeforeTick(*Connection){}
func (h NopConnHandler) OnAfterTick(*Connection){}