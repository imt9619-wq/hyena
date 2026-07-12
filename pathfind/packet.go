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
	packet.IDLevelChunk: _LevelChunkHandle{},
	packet.IDSubChunk: _SubChunkHandle{},
	packet.IDUpdateBlock: _UpdateBlockHandle{},
	packet.IDUpdateSubChunkBlocks: _UpdateSubChunkBlocksHandle{},
}

type _TextHandle struct{}
func (_TextHandle) handle(h *PathFindHandler, p packet.Packet){
	pk := p.(*packet.Text)
	isComeCmd := strings.Contains(pk.Message, "come")
	isCaller := strings.Contains(strings.ToLower(pk.SourceName), h.callerName)
	if !(isCaller && isComeCmd){
		return
	}
	tg, ok := h.c.EntityInWorld().NearByPlayerByXUID(pk.XUID)
	if !ok{
		return
	}
	h.executor.newEvent(EventGoalChanged, payLoadGoalChanged{goal: tg.Position})
	h.executor.newEvent(EventShouldMove, payLoadShouldMove{shouldMove: true})
}

type _LevelChunkHandle struct{}
func (_LevelChunkHandle) handle(h *PathFindHandler, p packet.Packet){
	pk := p.(*packet.LevelChunk)
	h.executor.newEvent(EventChunkChanged, payLoadChunkChanged{pos: pk.Position})
}

type _SubChunkHandle struct{}
func (_SubChunkHandle) handle(h *PathFindHandler, p packet.Packet){
	pk := p.(*packet.SubChunk)
	h.executor.newEvent(EventSubChunkChanged, payLoadSubChunkChanged{pos: pk.Position})
}

type _UpdateBlockHandle struct{}
func (_UpdateBlockHandle) handle(h *PathFindHandler, p packet.Packet){
	pk := p.(*packet.UpdateBlock)
	h.executor.newEvent(EventBlockChanged, payLoadBlockChanged{pos: pk.Position})
}

type _UpdateSubChunkBlocksHandle struct{}
func (_UpdateSubChunkBlocksHandle) handle(h *PathFindHandler, p packet.Packet){
	pk := p.(*packet.UpdateSubChunkBlocks)
	for _, bEntry := range pk.Blocks{
		h.executor.newEvent(EventBlockChanged, payLoadBlockChanged{pos: bEntry.BlockPos})
	}
	for _, bEntry := range pk.Extra{
		h.executor.newEvent(EventBlockChanged, payLoadBlockChanged{pos: bEntry.BlockPos})
	}
}