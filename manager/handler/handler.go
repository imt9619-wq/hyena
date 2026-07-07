package handler

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Handler receives connection lifecycle events for a server session.
type Handler interface {
	OnDisconnect(*Connection, string)
	OnJoin(*Connection)
	OnBeforeTick(*Connection)
	OnAfterTick(*Connection)
	OnNetworkStackLatency(*Context, *packet.NetworkStackLatency)
	OnLevelChunk(*Context, *packet.LevelChunk)
	OnSubChunk(*Context, *packet.SubChunk)
	OnNetworkChunkPublisherUpdate(*Context, *packet.NetworkChunkPublisherUpdate)
	OnChunkRadiusUpdated(*Context, *packet.ChunkRadiusUpdated)
	OnUpdateAttributes(*Context, *packet.UpdateAttributes)
	OnSetActorMotion(*Context, *packet.SetActorMotion)
	OnUpdateBlock(*Context, *packet.UpdateBlock)
	OnMovePlayer(*Context, *packet.MovePlayer)
	OnCorrectPlayerMovePrediction(*Context, *packet.CorrectPlayerMovePrediction)
	OnInventoryContent(*Context, *packet.InventoryContent)
	OnMobEquipment(*Context, *packet.MobEquipment)
	OnModalFormRequest(*Context, *packet.ModalFormRequest)
}

type NopConnHandler struct{}

var _ Handler = NopConnHandler{}

func (h NopConnHandler) OnModalFormRequest(*Context, *packet.ModalFormRequest){}
func (h NopConnHandler) OnMobEquipment(*Context, *packet.MobEquipment){}
func (h NopConnHandler) OnInventoryContent(*Context, *packet.InventoryContent){}
func (h NopConnHandler) OnDisconnect(*Connection, string){}
func (h NopConnHandler) OnJoin(*Connection){}
func (h NopConnHandler) OnBeforeTick(*Connection){}
func (h NopConnHandler) OnAfterTick(*Connection){}
func (h NopConnHandler) OnCorrectPlayerMovePrediction(*Context, *packet.CorrectPlayerMovePrediction){}
func (h NopConnHandler) OnNetworkStackLatency(*Context, *packet.NetworkStackLatency){}
func (h NopConnHandler) OnLevelChunk(*Context, *packet.LevelChunk){}
func (h NopConnHandler) OnSubChunk(*Context, *packet.SubChunk){}
func (h NopConnHandler) OnNetworkChunkPublisherUpdate(*Context, *packet.NetworkChunkPublisherUpdate){}
func (h NopConnHandler) OnChunkRadiusUpdated(*Context, *packet.ChunkRadiusUpdated){}
func (h NopConnHandler) OnUpdateAttributes(*Context, *packet.UpdateAttributes){}
func (h NopConnHandler) OnSetActorMotion(*Context, *packet.SetActorMotion){}
func (h NopConnHandler) OnUpdateBlock(*Context, *packet.UpdateBlock){}
func (h NopConnHandler) OnMovePlayer(*Context, *packet.MovePlayer){}