package utils

import (
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
