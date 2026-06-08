package blockmap

import (
	_ "github.com/df-mc/dragonfly/server/block"
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

func (b *BlockMap) GetBlockModel(pos mgl32.Vec3, layer uint8) (model world.BlockModel, exist bool) {
	model = nil
	exist = false
	if layer != 1{
		return 
	}

	chunkPos:= Mgl32ToWorldChunkPos(pos)
	chunk, ok := b.chunkMap[chunkPos]
	chunkRange := chunk.Range()
	if !(ok && chunkRange.Max() >= int(Float32Floor(pos[1])) && chunkRange.Min() <= int(Float32Floor(pos[1]))) {
		return
	}

	subChunk := chunk.SubChunk(chunk.SubIndex(int16(pos.Y())))
	x := byte(LastFourBit(int32(pos[0])))
	y := byte(LastFourBit(int32(pos[1])))
	z := byte(LastFourBit(int32(pos[2])))
	rid := subChunk.Block(x, y, z, layer)

	block, ok := world.BlockByRuntimeID(rid)
	if !ok{
		return
	}
	return block.Model(), true
}