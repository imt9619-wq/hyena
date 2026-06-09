package blockmap

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type palette struct {
	idToName  map[int16]string
	idToModel map[int16]world.BlockModel
}

func newPalette(items []protocol.ItemEntry) *palette {
	p := &palette{
		idToName:  make(map[int16]string, len(items)),
		idToModel: make(map[int16]world.BlockModel, len(items)),
	}
	for _, item := range items {
		p.idToName[item.RuntimeID] = item.Name

		if block, ok := world.BlockByRuntimeID(uint32(item.RuntimeID)); ok {
			if name, _ := block.EncodeBlock(); name == item.Name {
				fmt.Printf("successful match: %s", name)
				p.idToModel[item.RuntimeID] = block.Model()
				continue
			}
		}
		// Fallback: default world by name
		if block, ok := world.BlockByName(item.Name, nil); ok {
			p.idToModel[item.RuntimeID] = block.Model()
		}
	}
	return p
}