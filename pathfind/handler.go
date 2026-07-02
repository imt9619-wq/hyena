package pathfind

import (
	"fmt"

	"github.com/imt9619-wq/hyena/manager/handler"
)

type Handler struct {
	handler.NopConnHandler
}

func (h Handler) OnJoin(c *handler.Connection) {
	fmt.Printf("%s has joined the server: %s\n", c.IdentityData().DisplayName, c.RemoteAddr())
	c.StartRunning(false)
	c.StartSneaking(false)
	c.SetYaw(-40)
}

func (h Handler) OnDisconnect(c *handler.Connection, reason string) {
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
}
