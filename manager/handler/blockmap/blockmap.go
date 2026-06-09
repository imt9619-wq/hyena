package blockmap

import (
	"fmt"

	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// BlockMap is a map that hold *chunk.Chunk, the chunk inside the map
// should only be the chunk inside a player render distance, BlockMap
// is not safe to be used by mutiple gorotuines
type BlockMap struct {
	chunkMap map[world.ChunkPos]*chunk.Chunk
	chunkRadius int32
	chunkCentre world.ChunkPos
}

func NewBlockMap(conn *minecraft.Conn) *BlockMap {
	bm := &BlockMap{
		chunkRadius: 15,
	}
	bm.chunkCentre = Mgl32ToWorldChunkPos(conn.GameData().PlayerPosition)
	bm.chunkMap = make(map[world.ChunkPos]*chunk.Chunk, radiusToChunkCount(bm.chunkRadius))
	return bm
}

// When a player moved to a new chunk, chunk outside of their render 
// chunk distance will be deleted, when they get back in to the deleted 
// chunk(unloaded chunk), a levelchunk packet of that chunk should be 
// received to load back the chunk
func (b *BlockMap) UpdateChunkCentre(pos mgl32.Vec3) {
	chunkCentre := Mgl32ToWorldChunkPos(pos)
	if b.chunkCentre == chunkCentre{
		return
	}
	b.chunkCentre = chunkCentre
	b.RefreshMapWithRenderDistance() 
}

func (b *BlockMap) RefreshMapWithRenderDistance() {
	seCor, nwCor := getRenderedChunkFlame(b.chunkCentre, b.chunkRadius)
	for chunk := range b.chunkMap{
		if !isRenderedChunk(chunk, seCor, nwCor) {
			delete(b.chunkMap, chunk)
		}
	}
}

func (b *BlockMap) UpdateChunkRadius(r int32) {
	b.chunkRadius = r
}

func (b *BlockMap) InsertLevelChunk(pk *packet.LevelChunk) {
	airRID, _ := chunk.StateToRuntimeID("minecraft:air", nil)
	dim, _ := world.DimensionByID(int(pk.Dimension))
	chunk, err := chunk.NetworkDecode(airRID, pk.RawPayload, int(pk.SubChunkCount), dim.Range())
	if err != nil{
		fmt.Printf("Error when networkdecode chunk: %s", err)
		return
	}
	b.insertChunk(ProtocolPosToWorldPos(pk.Position), chunk)
}

func (b *BlockMap) insertChunk(pos world.ChunkPos, chunk *chunk.Chunk) {
	seCor, nwCor := getRenderedChunkFlame(b.chunkCentre, b.chunkRadius)
	if !isRenderedChunk(pos, seCor, nwCor) {
		return
	}
	b.chunkMap[pos] = chunk
}

func (b *BlockMap) SetBlock(pos protocol.BlockPos, layer uint8, block uint32) {
	chunkPos := ProtocolPosToWorldChunkPos(pos)
	chunk, ok := b.chunkMap[chunkPos]
	if !ok{
		return
	}
	x := uint8(LastFourBit(pos.X()))
	y := int16(pos[1])
	z := uint8(LastFourBit(pos.Z()))
	chunk.SetBlock(x, y, z, layer, block)
}

func (b *BlockMap) GetBlockModel(pos cube.Pos, layer uint8) (model world.BlockModel, exist bool) {
	model = nil
	exist = false
	if layer > 1{
		return 
	}

	chunkPos:= CubePosToChunkPos(pos)
	c, ok := b.chunkMap[chunkPos]
	if !ok {
		return
	}

	localX := uint8(pos[0]) & 0xF
	localZ := uint8(pos[2]) & 0xF
	worldY := int16(pos[1])

	rid := c.Block(localX, worldY, localZ, layer)

	block, ok := world.BlockByRuntimeID(rid)
	if !ok{
		return
	}
	return block.Model(), true
}