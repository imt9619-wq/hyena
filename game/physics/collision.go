package physics

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/utils"
)

type StateInWorld struct{
    Velocity mgl64.Vec3
    Position mgl64.Vec3
    AABB     cube.BBox

    world   *blockmap.BlockMap
    scratch *phyScratch
}

func NewStateInWorld(world *blockmap.BlockMap) *StateInWorld{
	s := &StateInWorld{}
	s.world = world
	s.scratch = newScratch()
	return s
}

// simState will use the given aabb, velocity, position and simulate the newState which the newState will 
// replace the old one
func (s *StateInWorld) SimState(){
	s.getOffset()
	s.simOffset()
}

func (s *StateInWorld) simOffset(){
	if utils.DeltaIsZero(s.Velocity){
		return
	}
	offsets := s.scratch.offsets.offsetArr()
	fmt.Println(s.Velocity)
	for axis, offset := range offsets{
		if offset == 0 && s.Velocity[axis] != 0 && len(s.scratch.offsets[axis].blocks) == 0{
			s.Velocity[axis] = 0
		}
	}
	fmt.Println(s.Velocity)	
	var radio float64
	deltas := s.Velocity
	for axis, minRadio := range utils.MinOffset(offsets, deltas){
		radio = minRadio
		if radio != 1{
			s.Velocity[axis] = 0	
		}	
	}
	fmt.Println(s.Position.Add(deltas.Mul(radio)))
	s.Position = s.Position.Add(deltas.Mul(radio))
}

func (s *StateInWorld) getOffset(){
	if utils.DeltaIsZero(s.Velocity){
		return
	}
	bm := s.world
	s.scratch.offsets.reset(s.Velocity)
	xFace, yFace, zFace := cube.FaceWest, cube.FaceDown, cube.FaceNorth
	if s.Velocity[0] > 0 {
		xFace = cube.FaceEast
	}
	if s.Velocity[1] > 0 {
		yFace = cube.FaceUp
	}
	if s.Velocity[2] > 0 {
		zFace = cube.FaceSouth
	}
	for pos := range s.scratch.sweptBlockPositions(s.AABB, s.Velocity) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok {
			continue
		}
		blockBoxes := bboxes(model, pos, bm)
		xSolid ,ySolid, zSolid := model.FaceSolid(pos, xFace, bm), model.FaceSolid(pos, yFace, bm), 
								  model.FaceSolid(pos, zFace, bm)
		for _, blockBox := range blockBoxes {
			s.scratch.offsets.considerOffsets(s.AABB, blockBox, [3]bool{xSolid, ySolid, zSolid}, s.Velocity)
		}
	}
}
