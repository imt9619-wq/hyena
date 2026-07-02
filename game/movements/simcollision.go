package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/movements/physics"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (m *Movement) simCollision(){
	m.setBlockSource()
	pos, velocity := m.doNormalCollisionThenStepAssist()
	for axis, plane := range velocity{
		if axis != 1 && plane == 0 && m.velocity[axis] != 0{
			m.setFlag(packet.InputFlagHorizontalCollision)
			break
		}
	}
	m.pasteStateToMovements(pos, velocity)
}

func (m *Movement) setBlockSource(){
	if m.IsSneak(){
		tinyBBox := utils.TinyBBoxOnBBoxFace(utils.PlayerBBox(m.position), cube.FaceDown)
		if utils.BBoxIntersectsSolid(m.world, tinyBBox) && m.velocity[1] <= 0{
			m.velocity[1] = 0
			m.blockSource = EdgeBlockSource{BlockMap: m.world}
			return
		}
	}
	m.blockSource = m.world
}

func (m *Movement) doStepAssist(op physics.OutPhyState) (pos, velocity mgl64.Vec3){
	pos, velocity = op.Position, op.Velocity
	if !m.onGround || utils.DeltaIsZero(m.velocity) || m.onClimb{
		return
	}

	var stepHeight float64 = 0
	var ceilHeight float64 = 1000
	var walkStairVelocity mgl64.Vec3
	for axis, plane := range velocity{
		if plane == 0 && m.velocity[axis] != 0 && axis != 1{
			walkStairVelocity[axis] = m.velocity[axis]
		}else{
			walkStairVelocity[axis] = 0
		}
	}
	var veloLen float64 = 1
	zeroVelo := mgl64.Vec3{}
	if walkStairVelocity != zeroVelo{
		veloLen = walkStairVelocity.Len()
	}
	pBBoxInStairs := m.bboxFunc(op.Position)
	pBBoxInStairs = pBBoxInStairs.Extend(walkStairVelocity.Mul(utils.ProbeOffset/veloLen))
	pBBoxInStairs = pBBoxInStairs.ExtendTowards(cube.FaceUp, MaxStepHeight)

	for _, blockBox := range utils.SweptBBoxInBBox(pBBoxInStairs, m.world){
		if pBBoxInStairs.IntersectsWith(blockBox){
			if blockBox.Min()[1] >= op.AABB.Max()[1]{
				ceilHeight = min(ceilHeight, blockBox.Min()[1]-op.AABB.Max()[1])
			}else if blockBox.Max()[1] >= op.AABB.Min()[1]{
				stepHeight = max(stepHeight, blockBox.Max()[1]-op.AABB.Min()[1])
			}				
		}
	}
	if stepHeight > MaxStepHeight || stepHeight == 0 || ceilHeight < stepHeight{
		return
	}
	// jump cancel
	velocityAfterStair := m.velocity
	if m.Space.Pressed && m.velocity[1] == JumpSpeed && stepHeight >= JumpSpeed{
		velocityAfterStair[1] = 0
	}
	
	stepOp := m.simAState(m.position.Add(mgl64.Vec3{0, stepHeight, 0}), velocityAfterStair)
	if stepOp.Position.Sub(m.position).Len() <= pos.Sub(m.position).Len(){
		return
	}
	return stepOp.Position, stepOp.Velocity
}

func (m *Movement) pasteStateToMovements(pos, velocity mgl64.Vec3){
	m.velocity = velocity
	m.position = pos
	if mgl64.FloatEqualThreshold(m.position[1], float64(m.world.Dimension().Range()[0]), utils.Negligible){
		m.position[1] = float64(m.world.Dimension().Range()[0])
	}
}

func (m *Movement) doNormalCollisionThenStepAssist() (mgl64.Vec3, mgl64.Vec3){
	return m.doStepAssist(m.simAState(m.position, m.velocity))
}

func (m *Movement) simAState(pos, velocity mgl64.Vec3) physics.OutPhyState{
	pos = utils.RoundVecTo5Decimal(pos)
	if eb, ok := m.blockSource.(EdgeBlockSource); ok{
		eb.pPos = pos
		eb.velocity = velocity
		m.blockSource = eb 
	}
	out := m.stateInWorld.SimState(physics.InPhyState{
		Position: pos,
		Velocity: velocity,
		BBoxFunc: m.bboxFunc,
		BlockSource: m.blockSource,
	})
	return out
}

// EdgeBlockSource is used instead of blockmap when the player is sneaking and is onGround, this type 
// will provide an extra BBox for some pos as boundaries to make sure the player wouldnt walk off edge when sneaking, 
// might be a bit more work than to just do the edge case in the collision logic, but it makes the code clearer in some way 
type EdgeBlockSource struct{
	*blockmap.BlockMap
    pPos     mgl64.Vec3
    velocity mgl64.Vec3
}

func (e EdgeBlockSource) BlockModel(pos cube.Pos, layer uint8) (model world.BlockModel, exist bool){
	blockUnderPos := pos.Sub(cube.Pos{0, 1, 0})
	model, exist = e.BlockMap.BlockModel(pos, layer)
	if model == nil{
		return
	}
	if int(math.Floor(e.pPos[1] - MaxStepHeight)) != blockUnderPos[1] || layer != 0{
		return
	}
	boundaryOffset := utils.PlayerWidth - utils.ProbeOffset
	ebboxMod := EdgeBBoxBlockModel{bboxs: model.BBox(pos, e.BlockMap)}
	blockUnderDiff := blockUnderPos.Vec3().Sub(pos.Vec3())
	underMod, _ := e.BlockMap.BlockModel(blockUnderPos, layer)
	if underMod == nil{
		return
	}
	underBox := underMod.BBox(blockUnderPos, e.BlockMap)
	// only bbox with Y in range of player PositionY-MaxStepHeight to PositionY can push the boundary
	sameY := func (self cube.BBox) bool{
		self = self.Translate(pos.Vec3())
		return self.Max()[1] <= e.pPos[1] && self.Max()[1] >= e.pPos[1] - MaxStepHeight
	}
	for axis, plane := range e.velocity{
		minOffsetBBox := utils.Box(mgl64.Vec3{}, mgl64.Vec3{}.Add(mgl64.Vec3{1, 1, 1})) // the boundary bbox we are going to add
		if axis == 1 || plane == 0{
			continue
		}
		pushedBBoxOnFace := false
		underOnNearbyFace := utils.FaceOnDeltaAxis(e.velocity, axis)
		pushBy := func (nearby cube.BBox) float64{
			var pushBy float64
			switch underOnNearbyFace{
			case cube.FaceEast, cube.FaceSouth:
				if nearby.Min()[axis] - minOffsetBBox.Min()[axis] - utils.ProbeOffset >= 0{
					return 0
				}
				pushBy = nearby.Max()[axis] + boundaryOffset - minOffsetBBox.Min()[axis] 
			default:
				if minOffsetBBox.Max()[axis] - nearby.Max()[axis] - utils.ProbeOffset >= 0{
					return 0
				}
				pushBy = minOffsetBBox.Max()[axis] - (nearby.Min()[axis] - boundaryOffset)
			}
			return utils.RoundFloat(pushBy, 3)
		}
		push := func (nearby cube.BBox) bool{
			pushBy := pushBy(nearby)
			if (minOffsetBBox.Max()[axis] - minOffsetBBox.Min()[axis]) <= pushBy{
				// boundary bbox is not in this blockUnderCube 
				return false
			}
			if pushBy > 0{
				minOffsetBBox = minOffsetBBox.ExtendTowards(underOnNearbyFace.Opposite(), -pushBy)
				pushedBBoxOnFace = true
			}
			return true
		}
		nearbyCube := cubePosDiffWithFace(underOnNearbyFace.Opposite())
		nearbyMod, _ := e.BlockMap.BlockModel(blockUnderPos.Add(nearbyCube), layer)
		if nearbyMod == nil{
			continue
		}
		nearbyDiff := nearbyCube.Vec3().Add(blockUnderDiff)
		for _, bbox := range nearbyMod.BBox(pos.Add(cube.PosFromVec3(nearbyDiff)), e.BlockMap){
			bbox = bbox.Translate(nearbyDiff)
			if !sameY(bbox){
				continue
			}
			if !push(bbox){
				return 
			}
		}
		for _, underBBox := range underBox{
			underBBox = underBBox.Translate(blockUnderDiff)
			if !sameY(underBBox){
				continue
			}
			if !push(underBBox){
				return
			}
		}
		if !pushedBBoxOnFace{
			minOffsetBBox = minOffsetBBox.ExtendTowards(underOnNearbyFace.Opposite(), -boundaryOffset)
		}
		canReach := reachable(utils.PlayerSneakBBox(e.pPos.Sub(pos.Vec3())), minOffsetBBox, axis, e.velocity)
		if canReach{
			if e.velocity[(axis+2)%4] != 0{
				minOffsetBBox = minOffsetBBox.ExtendTowards(utils.FaceOnDeltaAxis(e.velocity, (axis+2)%4), -boundaryOffset)
			}
			ebboxMod.bboxs = append(ebboxMod.bboxs, minOffsetBBox)
		}
	}
	return ebboxMod, exist
}

func cubePosDiffWithFace(faces ...cube.Face) cube.Pos{
	cPos := cube.Pos{}
	for _, face := range faces{
		switch face{
		case cube.FaceDown:
			cPos = cPos.Add(cube.Pos{0, -1, 0})
		case cube.FaceNorth:
			cPos = cPos.Add(cube.Pos{0, 0, -1})
		case cube.FaceEast:
			cPos = cPos.Add(cube.Pos{1, 0, 0})
		case cube.FaceSouth:
			cPos = cPos.Add(cube.Pos{0, 0, 1})
		case cube.FaceWest:
			cPos = cPos.Add(cube.Pos{-1, 0, 0})
		case cube.FaceUp:
			cPos = cPos.Add(cube.Pos{0, 1, 0})
		default:
		}
	}
	return cPos
}

type EdgeBBoxBlockModel struct{
	bboxs []cube.BBox
}

func (em EdgeBBoxBlockModel) BBox(cube.Pos, world.BlockSource) []cube.BBox{
	return em.bboxs
}

func (em EdgeBBoxBlockModel) FaceSolid(cube.Pos,cube.Face, world.BlockSource) bool{
	return false
}


func reachable(self, nearby cube.BBox, axis int, delta mgl64.Vec3) (reachable bool){	
	reachable = false
	var offset float64
	if delta[axis] == 0{
		return
	}
	if delta[axis] > 0 && self.Max()[axis] <= nearby.Min()[axis]+utils.Negligible {
		offset = min(nearby.Min()[axis] - self.Max()[axis], delta[axis])
	}else if delta[axis] < 0 && self.Min()[axis] >= nearby.Max()[axis]-utils.Negligible {
		offset = max(nearby.Max()[axis] - self.Min()[axis], delta[axis])
	}else{
		return
	}
	var radio float64 = 1
	if delta[axis] != 0{
		radio = offset/delta[axis]
	}
	if mgl64.FloatEqualThreshold(offset, 0, 0.1){
		offset = 0
	}
	if !utils.OutOfPlane(self.Translate(delta.Mul(radio)), nearby, axis){
		reachable = true
	}
	return
}
