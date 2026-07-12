package pathfind

import (
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type packetHandler interface{
	handle(h *PathFindHandler, pk packet.Packet)
}

var packetToPacketHandler = map[uint32]packetHandler{
	packet.IDText: _TextHandle{},
	packet.IDChunkRadiusUpdated: _ChunkRadiusUpdatedHandle{},
	packet.IDNetworkChunkPublisherUpdate: _NetworkChunkPublisherUpdate{},
	packet.IDLevelChunk: _LevelChunk{},
	packet.IDSubChunk: _SubChunk{},
	packet.IDUpdateBlock: _UpdateBlock{},
	packet.IDUpdateSubChunkBlocks: _UpdateSubChunkBlocks{},
}

type _TextHandle struct{}
func (_TextHandle) handle(h *PathFindHandler, p packet.Packet){
	pk := p.(*packet.Text)
	if !strings.Contains(strings.ToLower(pk.SourceName), h.callerName){
		return
	}

}

type _ChunkRadiusUpdatedHandle struct{}
func (_ChunkRadiusUpdatedHandle) handle(h *PathFindHandler, p packet.Packet){}
type _NetworkChunkPublisherUpdate struct{}
func (_NetworkChunkPublisherUpdate) handle(h *PathFindHandler, p packet.Packet){}
type _LevelChunk struct{}
func (_LevelChunk) handle(h *PathFindHandler, p packet.Packet){}
type _SubChunk struct{}
func (_SubChunk) handle(h *PathFindHandler, p packet.Packet){}
type _UpdateBlock struct{}
func (_UpdateBlock) handle(h *PathFindHandler, p packet.Packet){}
type _UpdateSubChunkBlocks struct{}
func (_UpdateSubChunkBlocks) handle(h *PathFindHandler, p packet.Packet){}