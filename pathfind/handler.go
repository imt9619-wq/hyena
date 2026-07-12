package pathfind

import (
	"fmt"

	"github.com/imt9619-wq/hyena/dbgshape"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/imt9619-wq/hyena/manager/handler/form"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type PathFindHandler struct {
	handler.NopConnHandler
	shape *dbgshape.ShapeCache
}

func NewPathFindHandler() *PathFindHandler{
	return &PathFindHandler{
		shape: dbgshape.NewShapeCache(),
	}
}

func (h *PathFindHandler) OnForm(ctx *handler.Context, f form.Form){
	clicked := false
	clickIfTitle := func (title, button string) bool{
		f, ok := f.(*form.Menu)
		if !ok{
			return false
		}
		if form.Resendable(f.Title(), title){
			return f.PressButton(button)
		}
		return false
	}
	clicked = clicked || clickIfTitle("server selector", "lobby")
	clicked = clicked || clickIfTitle("lobby", "lobby0")
	if clicked{
		fmt.Printf("Clicked button on %s\n", f.Title())
	}
}

func (h *PathFindHandler) OnJoin(c *handler.Connection){
	fmt.Printf("%s has joined the server: %s\n", c.IdentityData().DisplayName, c.RemoteAddr())
	c.SetYaw(-90)
}

func (h *PathFindHandler) OnBeforeTick(c *handler.Connection){
	if c.GameState().GStick() == 100{
		c.GameState().Inventory().SetHoldSlot(4)
		c.GameState().Exec(func(q *game.Qx) {
			c.GameState().Inputs().RightClick.Pressed = true
			c.GameState().Inputs().RightClick.PressOnce = true
		})
	}
}

func (h *PathFindHandler) OnDisconnect(c *handler.Connection, reason string){
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
	h.shape.Close()
}

func (h *PathFindHandler) OnPacket(ctx *handler.Context, pk packet.Packet){}