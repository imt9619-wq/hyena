package physics

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/utils"
)

type StateInWorld struct{
    Velocity mgl64.Vec3
    Position mgl64.Vec3
    AABB     cube.BBox

    moveVector [3]int
    world      *blockmap.BlockMap
    scratch    *phyScratch
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
	s.setMoveVector()
	s.getOffset()
	s.simOffset()
}

func (s *StateInWorld) setMoveVector(){
	for axis, plane := range s.Velocity{
		if plane == 0{
			s.moveVector[axis] = 0
		}else if plane > 0{
			s.moveVector[axis] = 1
		}else{
			s.moveVector[axis] = -1
		}
	}
}

func (s *StateInWorld) simOffset(){
	if utils.DeltaIsZero(s.Velocity){
		return
	}
	offsets := s.scratch.offsets.offsetArr()
	for axis := range s.Velocity{
		if s.isHittingWallOnAxis(axis){
			s.Velocity[axis] = 0
		}
	}

	var radio float64
	deltas := s.Velocity
	for axis, minRadio := range utils.MinOffset(offsets, deltas){
		radio = minRadio
		if radio != 1{
			s.Velocity[axis] = 0	
		}	
	}
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

func (s StateInWorld) isHittingWallOnAxis(axis int) bool{
	if axis == 1 || s.Velocity[axis] == 0{
		return false
	}
	halfHori := utils.HoriProbeOffset / 2
	axisAABBpos := s.Position
	axisAABBpos[axis] += (utils.PlayerWidth/2 + halfHori) * float64(s.moveVector[axis])
	return s.bboxIntersectsSolid(cube.Box(
			axisAABBpos[0]-halfHori,
			axisAABBpos[1]+utils.PlayerHeight,
			axisAABBpos[2]-halfHori,
			axisAABBpos[0]+halfHori,
			axisAABBpos[1],
			axisAABBpos[2]+halfHori,
			).Stretch(cube.Axis((axis%2)+1), utils.PlayerWidth/2-halfHori))
}

func (s *StateInWorld) bboxIntersectsSolid(pBBox cube.BBox) bool {
	bm := s.world
	for pos := range blockPositionsInBBox(pBBox) {
		model, ok := bm.BlockModel(pos, 0)
		if !ok {
			continue
		}
		for _, blockBox := range bboxes(model, pos, bm) {
			if pBBox.IntersectsWith(blockBox) {
				return true
			}
		}
	}
	return false
}
