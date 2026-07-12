package utils

import (
	"iter"
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// networkOffset can be found at github.com\df-mc\dragonfly\server\player.(ptype).NetworkOffset()
const (
	PlayerHeight           = float64(1.8)
	PlayerSneakHeight      = float64(1.5)
	PlayerWidth            = float64(0.6)
	NetworkOffset          = float64(1.62)
	ProbeOffset            = float64(0.003)
	Negligible             = float64(0.003)
	Epsilon                = float64(0.0001)
)

func ProtocolPosToMgl32Vec3(pos protocol.BlockPos) mgl32.Vec3{
	return mgl32.Vec3{float32(pos[0]), float32(pos[1]), float32(pos[2])}
}

func Mgl32Vec3ToProtocolPos(pos mgl32.Vec3) protocol.BlockPos{
	return protocol.BlockPos{int32(math.Floor(float64(pos[0]))), int32(math.Floor(float64(pos[1]))), int32(math.Floor(float64(pos[2])))}
}

func Mgl32ToWorldChunkPos(pos mgl32.Vec3) world.ChunkPos {
	chunkPosX := ShiftBackFourBits(int32(pos[0]))
	chunkPosZ := ShiftBackFourBits(int32(pos[2]))
	return world.ChunkPos([]int32{chunkPosX, chunkPosZ})
}

func CubePosToChunkPos(pos cube.Pos) world.ChunkPos {
	chunkPosX := ShiftBackFourBits(int32(pos[0]))
	chunkPosZ := ShiftBackFourBits(int32(pos[2]))
	return world.ChunkPos([]int32{chunkPosX, chunkPosZ})
}

func Float32Floor(x float32) float64 {
	return math.Floor(float64(x))
}

func ShiftBackFourBits(x int32) int32 {
	return x >> 4
}

func ProtocolBlockPosAdd(pos1, pos2 protocol.BlockPos) protocol.BlockPos{
	return protocol.BlockPos{pos1[0]+pos2[0], pos1[1]+pos2[1], pos1[2]+pos2[2]}
}

func SubChunkPosWithOffset(pos protocol.SubChunkPos, offset protocol.SubChunkOffset) protocol.SubChunkPos {
	return protocol.SubChunkPos{pos[0] + int32(offset[0]), pos[1] + int32(offset[1]), pos[2] + int32(offset[2])}
}

func SubChunkPosToChunkPos(pos protocol.SubChunkPos) protocol.ChunkPos{
	return protocol.ChunkPos{pos[0], pos[2]}
}

func ProtocolCPosToWorldCPos(pPos protocol.ChunkPos) world.ChunkPos {
	return world.ChunkPos([]int32{pPos.X(), pPos.Z()})
}

func RadiusToChunkCount(r int32) int32 {
	return int32(math.Pow(float64(r*2+1), 2))
}

func ProtocolPosToWorldChunkPos(protocolPos protocol.BlockPos) world.ChunkPos {
	return Mgl32ToWorldChunkPos(ProtocolPosToMgl32Vec3(protocolPos))
}

func LastFourBit(x int32) int32 {
	return x & 0x0F
}

func BBoxIntersectsSolid(bs BlockSourse, pBBox cube.BBox) bool {
	for _, blockBox := range SweptBBoxInBBox(pBBox, bs) {
		if pBBox.IntersectsWith(blockBox) {
			return true
		}
	}
	return false
}

func SweptBBoxInBBox(bbox cube.BBox, bs BlockSourse) iter.Seq2[cube.Pos, cube.BBox]{
	return func(yield func(cube.Pos, cube.BBox) bool) {
		for pos := range blockPositionsInBBox(bbox) {
			model, _ := bs.BlockModel(pos, 0)
			if model == nil {
				continue
			}
			for _, bbox := range BBoxes(model, pos, bs){
				if !yield(pos, bbox){
					return 
				}
			}
			
		}
	}
}

func blockPositionsInBBox(bbox cube.BBox) iter.Seq[cube.Pos]{
	min := bbox.Min()
	max := bbox.Max()
	
	return func(yield func(cube.Pos) bool) {
		for x := int(math.Floor(min[0])); x <= int(math.Floor(max[0])); x++ {
			for y := int(math.Floor(min[1])); y <= int(math.Floor(max[1])); y++ {
				for z := int(math.Floor(min[2])); z <= int(math.Floor(max[2])); z++ {
					if !yield(cube.Pos{x, y, z}){
						return 
					}
				}
			}
		}
	}
}

func BBoxes(model world.BlockModel, pos cube.Pos, s world.BlockSource) []cube.BBox{
	blockBoxes := model.BBox(pos, s)
	for i, bbox := range blockBoxes{
		blockBoxes[i] = bbox.Translate(pos.Vec3())
	}
	return blockBoxes
}

type AxisFace [3]cube.Face
func DeltaAxisFace(deltas mgl64.Vec3) AxisFace{
	a := AxisFace{cube.FaceWest, cube.FaceDown, cube.FaceNorth}
	if deltas[0] > 0 {
		a[0] = cube.FaceEast
	}
	if deltas[1] > 0 {
		a[1] = cube.FaceUp
	}
	if deltas[2] > 0 {
		a[2] = cube.FaceSouth
	}
	return a
}

func BlockInBBox(bbox cube.BBox, bs world.BlockSource) iter.Seq2[cube.Pos, world.Block]{
	return func(yield func(cube.Pos, world.Block) bool) {
		for blockPos := range blockPositionsInBBox(bbox){
			if !yield(blockPos, bs.Block(blockPos)){
				return 
			}
		}
	}
}

func BBoxOnBBoxFaceWithThreshold(self cube.BBox, face cube.Face, threshold float64) cube.BBox{
	min, max := self.Min(), self.Max()
	switch face{
	case cube.FaceUp:
		min[1] = max[1]
		max[1] += threshold
	case cube.FaceDown:
		max[1] = min[1]
		min[1] -= threshold
	case cube.FaceNorth:
		max[2] = min[2]
		min[2] -= threshold
	case cube.FaceEast:
		min[0] = max[0]
		max[0] += threshold
	case cube.FaceSouth:
		min[2] = max[2]
		max[2] += threshold
	default:
		max[0] = min[0]
		min[0] -= threshold
	}
	return cube.Box(min[0], min[1], min[2], max[0], max[1], max[2])
}

func TinyBBoxOnBBoxFace(self cube.BBox, face cube.Face) cube.BBox{
	return BBoxOnBBoxFaceWithThreshold(self, face, ProbeOffset)
}

func FaceOnDeltaAxis(delta mgl64.Vec3, axis int) cube.Face{
	switch axis{
	case 0:
		if delta[axis] > 0{
			return cube.FaceEast
		}else{
			return cube.FaceWest
		}
	case 1:
		if delta[axis] > 0{
			return cube.FaceUp
		}else{
			return cube.FaceDown
		}
	default:
		if delta[axis] > 0{
			return cube.FaceSouth
		}else{
			return cube.FaceNorth
		}
	}
}

func Box(vec1, vec2 mgl64.Vec3) cube.BBox{
	return cube.Box(vec1[0],
					vec1[1],
					vec1[2],
					vec2[0],
					vec2[1],
					vec2[2],
					)
}
