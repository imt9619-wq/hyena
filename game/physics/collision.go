package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

func EntityCollision(e PhysicsEntity) OutPhyState{
	out := OutPhyState{
		Position: e.Position(),
		Velocity: e.Velocity(),
	}
	aabb := e.BBox()
	dx, dy, dz := out.Velocity[0], out.Velocity[1], out.Velocity[2]

	blocksIn := aabb.ExtendTowards(utils.FaceOnDeltaAxis(out.Velocity, 1), math.Abs(out.Velocity[1]))
	for _, bbox := range utils.SweptBBoxInBBox(blocksIn, e.World()){
		dy = aabb.YOffset(bbox, dy)	
	}
	aabb = aabb.Translate(mgl64.Vec3{0, dy})

	minX := func (){
		blocksIn = aabb.ExtendTowards(utils.FaceOnDeltaAxis(out.Velocity, 0), math.Abs(out.Velocity[0]))
		for _, bbox := range utils.SweptBBoxInBBox(blocksIn, e.World()){
			dx = aabb.XOffset(bbox, dx)
		}
		aabb = aabb.Translate(mgl64.Vec3{dx})
	}
	minZ := func (){
		blocksIn = aabb.ExtendTowards(utils.FaceOnDeltaAxis(out.Velocity, 2), math.Abs(out.Velocity[2]))
		for _, bbox := range utils.SweptBBoxInBBox(blocksIn, e.World()){
			dx = aabb.ZOffset(bbox, dx)
		}
		aabb = aabb.Translate(mgl64.Vec3{0, 0, dz})
	}

	if math.Abs(dz) > math.Abs(dx){
		minX()
		minZ()
	}else{
		minZ()
		minX()
	}
	out.Position = out.Position.Add(mgl64.Vec3{dx, dy, dz})
	if dy != out.Velocity[1]{
		out.Onground = true
		out.Velocity[1] = 0
	}
	if dx != out.Velocity[0]{
		out.Velocity[0] = 0
	}
	if dz != out.Velocity[2]{
		out.Velocity[2] = 0
	}
	return out
}