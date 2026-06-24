package blockmap

import (
	"bytes"
	"fmt"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// we assume the clientCache is disenbled, if not, error will occur
func (b *BlockMap) InsertSubChunk(pk *packet.SubChunk) {
	center := pk.Position
	dim, _ := world.DimensionByID(int(pk.Dimension))
	r := dim.Range()
	buf := bytes.NewBuffer(nil)
	for _, entry := range pk.SubChunkEntries {
		entryPos := utils.SubChunkPosWithOffset(center, entry.Offset)
		if _, ok := b.subChunkInQuery[pk.Dimension][utils.SubChunkPosToChunkPos(entryPos)][entryPos[1]]; !ok {
			continue
		}
		delete(b.subChunkInQuery[pk.Dimension][utils.SubChunkPosToChunkPos(entryPos)], entryPos[1])
		if len(b.subChunkInQuery[pk.Dimension][utils.SubChunkPosToChunkPos(entryPos)]) == 0{
			delete(b.subChunkInQuery[pk.Dimension], utils.SubChunkPosToChunkPos(entryPos))
		}
		if !(entry.Result == protocol.SubChunkResultSuccess || entry.Result == protocol.SubChunkResultSuccessAllAir) {
			continue
		}
		c, ok := b.chunkMap[world.ChunkPos{entryPos[0], entryPos[2]}]
		if !ok {
			continue
		}

		ind := uint8(entryPos[1]-int32(r[0]>>4))
		var sub *chunk.SubChunk
		if entry.Result == protocol.SubChunkResultSuccessAllAir {
			sub = chunk.NewSubChunk(airRID)
		}else{
			buf.Write(entry.RawPayload)
			s, err := decodeSubChunk(buf, c, &ind, chunk.NetworkEncoding)
			buf.Reset()
			if err != nil {
				fmt.Printf("Error when networkdecode subChunk: %s\n", err)
				continue
			}
			sub = s
		}
		c.Sub()[ind] = sub
	}
}

func (b *BlockMap) InsertLevelChunk(pk *packet.LevelChunk) {
	if !b.isRenderedChunk(world.ChunkPos(pk.Position)) {
		return
	}
	b.currentDim = pk.Dimension
	dim, ok := world.DimensionByID(int(pk.Dimension))
	if !ok{
		pk.Dimension = 0
	}
	r := dim.Range()
	chunkPos := utils.ProtocolCPosToWorldCPos(pk.Position)
	if pk.SubChunkCount == protocol.SubChunkRequestModeLimited ||
		pk.SubChunkCount == protocol.SubChunkRequestModeLimitless {		
		highest := r[1]>>4
		if pk.SubChunkCount == protocol.SubChunkRequestModeLimited{
			highest = int(pk.HighestSubChunk) + r[0]>>4
		}
		if _, ok := b.subChunkInQuery[pk.Dimension][pk.Position]; !ok{
			b.subChunkInQuery[pk.Dimension][pk.Position] = make(map[int32]struct{}, highest-r[0]>>4)
		}
		for i := r[0] >> 4; i <= highest; i++{
			b.subChunkInQuery[pk.Dimension][pk.Position][int32(i)] = struct{}{}
		}
		b.insertChunk(chunkPos, chunk.New(airRID, r))
		//b.insertChunk(chunkPos, chunk.New(world.DefaultBlockRegistry, r))
		return
	}

	c, err := chunk.NetworkDecode(airRID, pk.RawPayload, int(pk.SubChunkCount), r)
	//c, err := chunk.NetworkDecode(world.DefaultBlockRegistry, pk.RawPayload, int(pk.SubChunkCount), r)
	if err != nil {
		fmt.Printf("Error when networkdecode chunk: %s\n", err)
		return
	}
	b.insertChunk(chunkPos, c)
}

// noinspection ALL
//
//go:linkname decodeSubChunk github.com/df-mc/dragonfly/server/world/chunk.decodeSubChunk
func decodeSubChunk(buf *bytes.Buffer, c *chunk.Chunk, index *byte, e chunk.Encoding) (*chunk.SubChunk, error)
