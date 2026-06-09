package blockmap

import (
	"sort"

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

	statesByName := dragonflyStatesByName()
	nameVariant := make(map[string]int)

	sorted := append([]protocol.ItemEntry(nil), items...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RuntimeID < sorted[j].RuntimeID
	})

	for _, item := range sorted {
		p.idToName[item.RuntimeID] = item.Name

		// Bedrock's StartGame palette is a unified item+block registry. Runtime IDs
		// in chunks index into this list — NOT into Dragonfly's block-only
		// block_states.nbt array. BlockByRuntimeID(serverRID) is therefore wrong
		// (e.g. server RID for diamond_shovel ≠ dragonfly block at that index).
		//
		// With empty ItemEntry.Data, disambiguate same-name block states by counting
		// how many times this name has already appeared in palette (RuntimeID) order.
		name := normalizePaletteName(item.Name)
		idx := nameVariant[name]
		nameVariant[name]++

		var props map[string]any
		if states := statesByName[name]; idx < len(states) {
			props = states[idx]
		}
		if block, ok := lookupBlock(name, props); ok {
			p.idToModel[item.RuntimeID] = block.Model()
		}
	}

	return p
}

func (p *palette) model(rid int16) (world.BlockModel, bool) {
	model, ok := p.idToModel[rid]
	return model, ok
}

func lookupBlock(name string, props map[string]any) (world.Block, bool) {
	if block, ok := world.BlockByName(normalizePaletteName(name), props); ok {
		return block, true
	}
	if props != nil {
		return world.BlockByName(normalizePaletteName(name), nil)
	}
	return nil, false
}

func dragonflyStatesByName() map[string][]map[string]any {
	states := make(map[string][]map[string]any)
	for _, block := range world.Blocks() {
		name, props := block.EncodeBlock()
		states[name] = append(states[name], props)
	}
	return states
}

func normalizePaletteName(name string) string {
	switch name {
	case "minecraft:tallgrass":
		return "minecraft:short_grass"
	case "minecraft:grass":
		return "minecraft:grass_block"
	case "minecraft:flowing_water":
		return "minecraft:water"
	case "minecraft:flowing_lava":
		return "minecraft:lava"
	case "minecraft:item.campfire":
		return "minecraft:campfire"
	case "minecraft:item.crimson_door":
		return "minecraft:crimson_door"
	case "minecraft:item.flower_pot":
		return "minecraft:flower_pot"
	case "minecraft:item.jungle_door":
		return "minecraft:jungle_door"
	case "minecraft:item.brewing_stand":
		return "minecraft:brewing_stand"
	default:
		return name
	}
}
