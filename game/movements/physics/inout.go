package physics

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

type InPhyState struct {
	Velocity    mgl64.Vec3
    Position    mgl64.Vec3
    BBoxFunc    utils.BBoxFunc
    BlockSource utils.BlockSourse
}

// we are going to round off player position to the last five digit as the player might be stuck(rare but possible)
// if they got something like Z: 88.19999694824219 and is in front of a stair
func (s *StateInWorld) copyInPhyState(state InPhyState) {
	s.position = utils.RoundVecTo5Decimal(state.Position)
	s.aaBB = state.BBoxFunc(s.position)
	s.velocity = utils.RemoveDeltaEpsilon(state.Velocity)
	s.world = state.BlockSource
}

type OutPhyState struct {
	Velocity mgl64.Vec3
	Position mgl64.Vec3
	AABB     cube.BBox
}

func (s *StateInWorld) outPhyState() OutPhyState {
	return OutPhyState{
		Velocity: s.velocity,
		Position: s.position,
		AABB:     s.aaBB,
	}
}
