package pathfind

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/dbgshape"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/imt9619-wq/hyena/utils"
)

type Handler struct {
	handler.NopConnHandler
	shape *dbgshape.ShapeCache
}

func NewPathHandler() *Handler{
	return &Handler{
		shape: dbgshape.NewShapeCache(),
	}
}

func (h *Handler) OnJoin(c *handler.Connection) {
	fmt.Printf("%s has joined the server: %s\n", c.IdentityData().DisplayName, c.RemoteAddr())
	c.StartRunning(false)
	c.StartJumping(false)
	c.SetYaw(0)
}

func (h *Handler) OnDisconnect(c *handler.Connection, reason string) {
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
	h.shape.Close()
}

func (h *Handler) OnAfterTick(c *handler.Connection){
	h.writeShape(c)
}

func (h *Handler) writeShape(c *handler.Connection){
	pPos := c.GameState().Player().Position().Sub(mgl32.Vec3{0, float32(utils.NetworkOffset)})
	/*pBBox := utils.PlayerBBox(utils.Mgl32Vec3Tomgl64Vec3(pPos))
	h.shape.AddDeadlineShape(c.Conn, dbgshape.AShape{
		Shape: dbgshape.Box,
		StartPoint: utils.Mgl64Vec3Tomgl32Vec3(pBBox.Min()),
		EndPoint: utils.Mgl64Vec3Tomgl32Vec3(pBBox.Max()),
	}, time.Millisecond*100)*/
	h.shape.AddDeadlineShape(c.Conn, dbgshape.AShape{
		Shape: dbgshape.Line,
		StartPoint: pPos,
		EndPoint: pPos.Add(c.GameState().Player().Velocity()),
	}, time.Millisecond*100)
}