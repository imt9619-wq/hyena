package blockmap

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block"
	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)
var airRID uint32
func init() {
	world.DefaultBlockRegistry.Finalize()
	airRID = world.BlockRuntimeID(block.Air{})
}

// BlockMap holds *chunk.Chunk values for chunks within render distance.
// BlockMap is not safe for use by multiple goroutines.
type BlockMap struct {
	chunkMap    map[world.ChunkPos]*chunk.Chunk
	chunkRadius int32
	chunkCentre world.ChunkPos
	subChunkInQuery [3]map[protocol.ChunkPos]map[int32]struct{}
	currentDim int32
}

func NewBlockMap(conn *minecraft.Conn) *BlockMap {
	bm := &BlockMap{
		chunkRadius: 15,
	}
	bm.chunkCentre = Mgl32ToWorldChunkPos(conn.GameData().PlayerPosition)
	bm.chunkMap = make(map[world.ChunkPos]*chunk.Chunk, radiusToChunkCount(bm.chunkRadius))
	for i := range bm.subChunkInQuery{
		bm.subChunkInQuery[i] = make(
			map[protocol.ChunkPos]map[int32]struct{}, 
			radiusToChunkCount(bm.chunkRadius))
	}
	return bm
}

// When a player moved to a new chunk, chunk outside of their render
// chunk distance will be deleted, when they get back in to the deleted
// chunk(unloaded chunk), a levelchunk packet of that chunk should be
// received to load back the chunk
func (b *BlockMap) UpdateChunkCentre(pos mgl32.Vec3) {
	chunkCentre := Mgl32ToWorldChunkPos(pos)
	if b.chunkCentre == chunkCentre {
		return
	}
	b.chunkCentre = chunkCentre
}

func (b *BlockMap) RefreshMapWithRenderDistance() {
	for chunk := range b.chunkMap {
		if !b.isRenderedChunk(chunk) {
			delete(b.chunkMap, chunk)
		}
	}
	for i, subQuery := range b.subChunkInQuery{
		if i == int(b.currentDim){
			for subchunk := range subQuery{
				if !b.isRenderedChunk([2]int32{subchunk[0], subchunk[1]}) {
					delete(b.subChunkInQuery[i], subchunk)
				}
			}
			continue
		}
		clear(b.subChunkInQuery[i])
	}
	
}

func (b *BlockMap) UpdateChunkRadius(r int32) {
	b.chunkRadius = r
}

func (b *BlockMap) insertChunk(pos world.ChunkPos, chunk *chunk.Chunk) {
	if !b.isRenderedChunk(pos) {
		return
	}
	b.chunkMap[pos] = chunk
}

func (b *BlockMap) SetBlock(pos protocol.BlockPos, layer uint8, block uint32) {
	chunkPos := ProtocolPosToWorldChunkPos(pos)
	chunk, ok := b.chunkMap[chunkPos]
	if !ok {
		return
	}
	x := uint8(LastFourBit(pos.X()))
	y := int16(pos[1])
	z := uint8(LastFourBit(pos.Z()))
	chunk.SetBlock(x, y, z, layer, block)
}

// Block implements world.BlockSource.
func (b *BlockMap) Block(pos cube.Pos) world.Block {
	bl, _ := b.block(pos, 0)
	return bl
}

func (b *BlockMap) SubChunkInQuery()  [3]map[protocol.ChunkPos]map[int32]struct{}{
	return b.subChunkInQuery
}

func (b *BlockMap) block(pos cube.Pos, layer uint8) (bl world.Block, exist bool) {
	bl = nil
	exist = false
	if layer > 1 {
		return
	}

	chunkPos := CubePosToChunkPos(pos)
	c, ok := b.chunkMap[chunkPos]
	if !ok {
		fmt.Printf("Tried to query out of render distance blocks\n")
		bl = block.InvisibleBedrock{}
		return
	}

	localX := uint8(pos[0]) & 0xF
	localZ := uint8(pos[2]) & 0xF
	worldY := int16(pos[1])

	if !(c.Range()[0] <= int(worldY) && int(worldY) <= c.Range()[1]){
		fmt.Printf("Tried to query out of range blocks\n")
		bl, exist = world.BlockByRuntimeID(airRID)
		return
	}
	rid := c.Block(localX, worldY, localZ, layer)

	bl, exist = world.BlockByRuntimeID(rid)
	return
}

func (b *BlockMap) BlockModel(pos cube.Pos, layer uint8) (model world.BlockModel, exist bool) {
	model = nil
	exist = false
	block, exist := b.block(pos, layer)
	if !exist {
		return
	}
	return block.Model(), exist
}

func (b *BlockMap) isRenderedChunk(chunk [2]int32) bool {
	return b.chunkCentre[0]-b.chunkRadius <= chunk[0] && chunk[0] <= b.chunkCentre[0]+b.chunkRadius &&
	b.chunkCentre[1]-b.chunkRadius <= chunk[1] && chunk[1] <= b.chunkCentre[1]+b.chunkRadius
}