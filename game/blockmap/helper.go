package blockmap

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func ProtocolPosToMgl32Vec3(protocolPos protocol.BlockPos) mgl32.Vec3 {
	posInMgl32 := mgl32.Vec3([]float32{float32(protocolPos[0]), float32(protocolPos[1]), float32(protocolPos[2])})
	return posInMgl32
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

func ProtocolCPosToWorldCPos(pPos protocol.ChunkPos) world.ChunkPos {
	return world.ChunkPos([]int32{pPos.X(), pPos.Z()})
}

func getRenderedChunkFrame(chunkCentre world.ChunkPos, chunkRadius int32) (SouthEastCorner world.ChunkPos, NorthWestCorner world.ChunkPos) {
	SouthEastCorner = world.ChunkPos([]int32{chunkCentre.X() + chunkRadius, chunkCentre.Z() + chunkRadius})
	NorthWestCorner = world.ChunkPos([]int32{chunkCentre.X() - chunkRadius, chunkCentre.Z() - chunkRadius})
	return
}

func radiusToChunkCount(r int32) int32 {
	return int32(math.Pow(float64(r*2+1), 2))
}

func isRenderedChunk(chunk world.ChunkPos, seCornerChunk world.ChunkPos, nwCornerChunk world.ChunkPos) bool {
	isInXRange := (chunk.X() <= seCornerChunk.X() && chunk.X() >= nwCornerChunk.X())
	isInZRange := (chunk.Z() <= seCornerChunk.Z() && chunk.Z() >= nwCornerChunk.Z())
	return isInXRange && isInZRange
}

func ProtocolPosToWorldChunkPos(protocolPos protocol.BlockPos) world.ChunkPos {
	return Mgl32ToWorldChunkPos(ProtocolPosToMgl32Vec3(protocolPos))
}

func LastFourBit(x int32) int32 {
	return x & 0x0F
}