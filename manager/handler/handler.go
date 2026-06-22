package handler

import "fmt"

// Handler receives connection lifecycle events for a server session.
type Handler interface {
	OnDisconnect(*Connection, string)
	OnJoin(*Connection)
}

type DefaultHandler struct{}

func (h DefaultHandler) OnDisconnect(c *Connection, reason string) {
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
}

func (h DefaultHandler) OnJoin(c *Connection) {
	fmt.Printf("%s has joined the server: %s\n", c.IdentityData().DisplayName, c.RemoteAddr())
	c.StartRunning()
	//c.StartJumping()
}
