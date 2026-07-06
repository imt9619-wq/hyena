package physics

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)


type PhysicsEntity interface{
	Velocity() mgl64.Vec3
	Position() mgl64.Vec3
	BBox() cube.BBox
	World() utils.BlockSourse
}

type NopEntity struct{
	Pos, Vec mgl64.Vec3
	Bs utils.BlockSourse
	AAbb cube.BBox
}

func (n NopEntity) Velocity() mgl64.Vec3{
	return utils.RoundVecTo5Decimal(n.Vec)
}

func (n NopEntity) Position() mgl64.Vec3{
	return utils.RemoveDeltaEpsilon(n.Pos)
}

func (n NopEntity) World() utils.BlockSourse{
	return n.Bs
}

func (n NopEntity) BBox() cube.BBox{
	return n.AAbb
}

type OutPhyState struct {
	Velocity mgl64.Vec3
	Position mgl64.Vec3
}

