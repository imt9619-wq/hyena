package pathfind

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/dbgshape"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/imt9619-wq/hyena/utils"
)

func (h *PathFindHandler) writeShape(c *handler.Connection) {
	pPos := c.GameState().Player().Position().Sub(mgl32.Vec3{0, float32(utils.NetworkOffset)})
	/*pBBox := utils.PlayerBBox(utils.Mgl32Vec3Tomgl64Vec3(pPos))
	h.shape.AddDeadlineShape(c.Conn, dbgshape.AShape{
		Shape: dbgshape.Box,
		StartPoint: utils.Mgl64Vec3Tomgl32Vec3(pBBox.Min()),
		EndPoint: utils.Mgl64Vec3Tomgl32Vec3(pBBox.Max()),
	}, time.Millisecond*100)*/
	h.shape.AddDeadlineShape(c.Conn, dbgshape.AShape{
		Shape:      dbgshape.Line,
		StartPoint: pPos,
		EndPoint:   pPos.Add(c.GameState().Player().Velocity()),
	}, time.Millisecond*100)
}