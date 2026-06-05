package blockmap

import (
	"sync"

	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockMap maps server network runtime IDs (from StartGame + chunk palettes) to
// block names and Dragonfly BlockModels used for collision (BBox / FaceSolid).
type BlockMap struct {
	mu sync.RWMutex

	runtimeIDToName  map[int16]string
	runtimeIDToModel map[int16]world.BlockModel
}

func NewBlockMap(conn *minecraft.Conn) *BlockMap {
	bm := &BlockMap{
		runtimeIDToName:  make(map[int16]string, 1800),
		runtimeIDToModel: make(map[int16]world.BlockModel, 1200),
	}
	// Build synchronously so the map is ready before LevelChunk packets arrive.
	bm.registerItems(conn.GameData().Items)
	return bm
}

func (bm *BlockMap) registerItems(items []protocol.ItemEntry) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, item := range items {
		bm.runtimeIDToName[item.RuntimeID] = item.Name

		block, ok := world.BlockByName(item.Name, nil)
		if ok {
			bm.runtimeIDToModel[item.RuntimeID] = block.Model()
			continue
		}
		// Block exists on server but has no Dragonfly implementation — treat as full cube.
		bm.runtimeIDToModel[item.RuntimeID] = solidModel{}
	}
}

// Name returns the block identifier for a server runtime ID.
func (bm *BlockMap) Name(runtimeID int16) (string, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	name, ok := bm.runtimeIDToName[runtimeID]
	return name, ok
}

// Model returns the collision model for a server runtime ID.
// Use this after reading a uint32 from a decoded chunk (cast to int16).
func (bm *BlockMap) Model(runtimeID int16) (world.BlockModel, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	model, ok := bm.runtimeIDToModel[runtimeID]
	return model, ok
}

// ModelAt returns the model for a runtime ID read from chunk data.
func (bm *BlockMap) ModelAt(runtimeID uint32) world.BlockModel {
	model, ok := bm.Model(int16(runtimeID))
	if ok {
		return model
	}
	return solidModel{}
}

// solidModel is a 1x1x1 cube fallback for unknown or unimplemented blocks.
type solidModel struct{}

func (solidModel) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
}

func (solidModel) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return true
}
