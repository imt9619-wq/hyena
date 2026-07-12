package handler

import (
	"strings"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type EntInWorld struct{
	selfUUid      uuid.UUID
    mu            *sync.RWMutex
    ents          map[int64]*PlayerNearBy
	entsInWorld   map[uuid.UUID]*PlayerInWorld
    networkRadius uint32
	networkCentre mgl32.Vec3
}

type PlayerInWorld struct{
	Rid      int64
    UUID     uuid.UUID
    XUID     string
    Name     string
}

type PlayerNearBy struct {
	UUID     uuid.UUID
	XUID     string
	Name     string
	Rid      int64
    Position mgl32.Vec3
    Yaw      float32
    Pitch    float32
    OnGround bool
}

func newEntInWorld(conn *minecraft.Conn) *EntInWorld{
	id, err := uuid.Parse(conn.IdentityData().Identity)
	if err != nil{
		panic("Non uuid Identity.")
	}
	return &EntInWorld{
		mu: &sync.RWMutex{},
		ents: make(map[int64]*PlayerNearBy, 128),
		entsInWorld: make(map[uuid.UUID]*PlayerInWorld, 128),
		selfUUid: id,
		networkRadius: 128,
	}
}

func (ew *EntInWorld) syncNetworkChunk(r uint32, centre protocol.BlockPos){
	ew.mu.Lock()
	defer ew.mu.Unlock()
	ew.networkRadius = r
	ew.networkCentre = utils.ProtocolPosToMgl32Vec3(centre)
	for rid, p := range ew.ents{
		pos := p.Position
		pos[1] = ew.networkCentre[1]
		if pos.Sub(ew.networkCentre).Len() > float32(r){
			delete(ew.ents, rid)
		}
	}
}

func (ew *EntInWorld) handlePlayerList(pk *packet.PlayerList){
	ew.mu.Lock()
	defer ew.mu.Unlock()

	switch pk.ActionType{
	case packet.PlayerListActionAdd:
		for _, pEntry := range pk.Entries{
			if pEntry.UUID == ew.selfUUid{
				continue
			}
			if _, ok := ew.entsInWorld[pEntry.UUID]; !ok{
				ew.entsInWorld[pEntry.UUID] = &PlayerInWorld{
					Rid: -1,
					UUID: pEntry.UUID,
					Name: pEntry.Username,
					XUID: pEntry.XUID,
				}
			}
		}
	case packet.PlayerListActionRemove:
		for _, pEntry := range pk.Entries{
			if p, ok := ew.entsInWorld[pEntry.UUID]; ok{
				delete(ew.ents, p.Rid)
			}
			delete(ew.entsInWorld, pEntry.UUID)
		}
	}
}

func (ew *EntInWorld) handleAddPlayer(pk *packet.AddPlayer){
	rid := int64(pk.EntityRuntimeID)
	ew.mu.Lock()
	defer ew.mu.Unlock()
	pw, ok := ew.entsInWorld[pk.UUID]
	if !ok{		
		return
	}
	pw.Rid = rid
	if pk.UUID == ew.selfUUid{
		return
	}
	p, ok := ew.ents[rid]
	if !ok{
		p = &PlayerNearBy{}
	}
	p.XUID = pw.XUID
	p.Pitch = pk.Pitch
	p.Yaw = pk.Yaw
	p.Position = pk.Position
	p.UUID = pk.UUID
	p.Rid = rid
	p.Name = pk.Username
	ew.ents[rid] = p
}

func (ew *EntInWorld) movePlayer(pk *packet.MovePlayer){
	ew.mu.Lock()
	defer ew.mu.Unlock()
	p, ok := ew.ents[int64(pk.EntityRuntimeID)]
	if !ok{
		return
	}
	p.Position = pk.Position
	p.Pitch = pk.Pitch
	p.Yaw = pk.Yaw
	p.OnGround = pk.OnGround
}

func (ew *EntInWorld) NearByPlayerByName(name string) (PlayerNearBy, bool){
	name = strings.ToLower(name)
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	for _, p := range ew.ents{
		if strings.ToLower(p.Name) == name{
			return *p, true
		}
	}
	return PlayerNearBy{}, false
}

func (ew *EntInWorld) NearByPlayerByXUID(xuid string) (PlayerNearBy, bool){
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	for _, p := range ew.ents{
		if p.XUID == xuid{
			return *p, true
		}
	}
	return PlayerNearBy{}, false
}