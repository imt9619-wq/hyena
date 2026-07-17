package pathfind

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/physics"
	"github.com/imt9619-wq/hyena/utils"
)

type mode int
const(
	Normal mode = iota
	Sneak
)

func (e *pathRunExecutor) canReach(mode mode, start, end mgl64.Vec3) bool{
	ent := physics.NopEntity{
		Pos: start,
		Vec: end.Sub(start),
		Bs: e.world,
	}
	if mode == Sneak{
		ent.AAbb = utils.PlayerSneakBBox(start)
	}else{
		ent.AAbb = utils.PlayerBBox(start)
	}
	if cube.PosFromVec3(physics.EntityCollision(ent).Position) !=  cube.PosFromVec3(end){
		return false
	}
	return true
}