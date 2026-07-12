package handler

import (
	"strconv"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type EntInWorld struct{
	selfRid int64
    mu      *sync.RWMutex
    ents    map[int64]*PlayerNearBy
}

type PlayerNearBy struct {
	rid  int64
	uuid uuid.UUID
    xuid int
    name string
    pos  mgl32.Vec3
    yaw  float32
	pitch float32
	OnGround bool
}

func newEntInWorld(conn *minecraft.Conn) *EntInWorld{
	return &EntInWorld{
		mu: &sync.RWMutex{},
		ents: make(map[int64]*PlayerNearBy, 100),
		selfRid: conn.GameData().EntityUniqueID,
	}
}

func (ew *EntInWorld) handlePlayerList(pk *packet.PlayerList){
	ew.mu.Lock()
	defer ew.mu.Unlock()

	switch pk.ActionType{
	case packet.PlayerListActionAdd:
		for _, pEntry := range pk.Entries{
			if pEntry.EntityUniqueID == ew.selfRid{
				continue
			}
			p, ok := ew.ents[pEntry.EntityUniqueID]
			if !ok{
				p = &PlayerNearBy{}
			}
			id, err := strconv.Atoi(pEntry.XUID)
			if err != nil {
				delete(ew.ents, pEntry.EntityUniqueID)
				continue
			}
			p.xuid = id
			p.name = pEntry.Username
			p.uuid = pEntry.UUID
			p.rid = pEntry.EntityUniqueID
		}
	case packet.PlayerListActionRemove:
		for _, pEntry := range pk.Entries{
			delete(ew.ents, pEntry.EntityUniqueID)
		}
	}
}

func (ew *EntInWorld) movePlayer(pk *packet.MovePlayer){
	ew.mu.Lock()
	defer ew.mu.Unlock()
	p, ok := ew.ents[int64(pk.EntityRuntimeID)]
	if !ok{
		return
	}
	p.pos = pk.Position
	p.pitch = pk.Pitch
	p.yaw = pk.Yaw
	p.OnGround = pk.OnGround
}

func (ew *EntInWorld) PlayerByRid(rid int64) (*PlayerNearBy, bool){
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	p, ok := ew.ents[rid]
	return p, ok
}

func (ew *EntInWorld) PlayerByName(name string) (*PlayerNearBy, bool){
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	for _, p := range ew.ents{
		if p.name == name{
			return p, true
		}
	}
	return nil, false
}

func (ew *EntInWorld) PlayerByUUID(id uuid.UUID) (*PlayerNearBy, bool){
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	for _, p := range ew.ents{
		if p.uuid == id{
			return p, true
		}
	}
	return nil, false
}

func (ew *EntInWorld) PlayerByXUID(id int) (*PlayerNearBy, bool){
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	for _, p := range ew.ents{
		if p.xuid == id{
			return p, true
		}
	}
	return nil, false
}