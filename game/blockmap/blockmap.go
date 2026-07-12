package blockmap

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block"
	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap/hblock"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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
	packets    utils.PacketBuffer
}

func NewBlockMap(conn *minecraft.Conn, pks utils.PacketBuffer) *BlockMap {
	bm := &BlockMap{
		chunkRadius: 15,
		packets: pks,
	}
	bm.chunkCentre = utils.Mgl32ToWorldChunkPos(conn.GameData().PlayerPosition)
	bm.chunkMap = make(map[world.ChunkPos]*chunk.Chunk, utils.RadiusToChunkCount(bm.chunkRadius))
	for i := range bm.subChunkInQuery{
		bm.subChunkInQuery[i] = make(
			map[protocol.ChunkPos]map[int32]struct{}, 
			utils.RadiusToChunkCount(bm.chunkRadius))
	}
	return bm
}

func (b *BlockMap) Dimension() world.Dimension{
	dim, _ := world.DimensionByID(int(b.currentDim))
	return dim
}

// When a player moved to a new chunk, chunk outside of their render
// chunk distance will be deleted, when they get back in to the deleted
// chunk(unloaded chunk), a levelchunk packet of that chunk should be
// received to load back the chunk
func (b *BlockMap) UpdateChunkCentre(pos mgl32.Vec3) {
	chunkCentre := utils.Mgl32ToWorldChunkPos(pos)
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
	chunkPos := utils.ProtocolPosToWorldChunkPos(pos)
	chunk, ok := b.chunkMap[chunkPos]
	if !ok {
		return
	}
	x := uint8(utils.LastFourBit(pos.X()))
	y := int16(pos[1])
	z := uint8(utils.LastFourBit(pos.Z()))
	chunk.SetBlock(x, y, z, layer, block)
}

// Block implements world.BlockSource.
func (b *BlockMap) Block(pos cube.Pos) world.Block {
	bl, _ := b.block(pos, 0)
	return bl
}

func (b *BlockMap) RequestSubChunkInQuery(){
	for dim, query := range b.subChunkInQuery{
		queryl := len(query)
		if queryl == 0 {
			continue
		}
		dimen, _ := world.DimensionByID(dim)
		r := dimen.Range()
		offsets := make([]protocol.SubChunkOffset, 0, r.Height()>>4)
		var pos int32
		for cpos, chunkSub := range query{
			offsets = offsets[:0]
			for subPos := range chunkSub {
				pos = subPos
				break
			}
			for subPos := range chunkSub {
				offsets = append(offsets, [3]int8{0, int8(subPos - pos), 0})
			}
			b.packets.Append(&packet.SubChunkRequest{
				Dimension: int32(dim),
				Position:  protocol.SubChunkPos{cpos[0], pos, cpos[1]},
				Offsets:   offsets,
			})
		}
		
	}
}

func (b *BlockMap) block(pos cube.Pos, layer uint8) (bl world.Block, exist bool) {
	bl = nil
	exist = false
	if layer > 1 {
		return
	}

	chunkPos := utils.CubePosToChunkPos(pos)
	c, ok := b.chunkMap[chunkPos]
	if !ok {
		//fmt.Printf("Tried to query out of render distance blocks (Cpos: %v)\n", chunkPos)
		bl = block.InvisibleBedrock{}
		return
	}

	localX := uint8(pos[0]) & 0xF
	localZ := uint8(pos[2]) & 0xF
	worldY := int16(pos[1])

	if !(c.Range()[0] <= int(worldY) && int(worldY) <= c.Range()[1]){
		fmt.Printf("Tried to query out of range blocks Y: %d\n", worldY)
		bl, exist = world.BlockByRuntimeID(airRID)
		return
	}
	rid := c.Block(localX, worldY, localZ, layer)

	bl, exist = world.BlockByRuntimeID(rid)
	return
}

func (b *BlockMap) BlockModel(pos cube.Pos, layer uint8) (world.BlockModel, bool) {
	block, exist := b.block(pos, layer)
	if block == nil {
		return nil, exist
	}
	return block.Model(), exist
}

func (b *BlockMap) isRenderedChunk(chunk [2]int32) bool {
	return b.chunkCentre[0]-b.chunkRadius <= chunk[0] && chunk[0] <= b.chunkCentre[0]+b.chunkRadius &&
	b.chunkCentre[1]-b.chunkRadius <= chunk[1] && chunk[1] <= b.chunkCentre[1]+b.chunkRadius
}

func (b *BlockMap) Hblock(pos cube.Pos) hblock.Block {
	bl, _ := b.block(pos, 0)
	return hblock.WblockToHblock(bl)
}